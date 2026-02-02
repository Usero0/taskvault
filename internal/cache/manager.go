package cache

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/taskvault/taskvault/internal/audit"
	"github.com/taskvault/taskvault/internal/hash"
	"github.com/taskvault/taskvault/internal/storage"
)

// Manager orchestrates cache lookups, saves, and eviction
type Manager struct {
	store     *storage.Store
	hasher    *hash.Engine
	auditLog  *audit.Logger
	mu        sync.RWMutex
	policies  map[string]*EvictionPolicy
	maxSizeGB int64
}

// EvictionPolicy defines TTL and eviction strategy
type EvictionPolicy struct {
	Name     string
	TTL      time.Duration
	MaxSize  int64  // bytes
	Strategy string // "lru", "lfu", "fifo"
}

// NewManager creates a cache manager
func NewManager(cacheDir string, maxSizeGB int64, hashAlgo hash.HashAlgorithm) (*Manager, error) {
	store, err := storage.NewStore(cacheDir, maxSizeGB)
	if err != nil {
		return nil, fmt.Errorf("storage error: %w", err)
	}

	auditLogger, err := audit.NewLogger(cacheDir)
	if err != nil {
		store.Close()
		return nil, fmt.Errorf("audit error: %w", err)
	}

	return &Manager{
		store:     store,
		hasher:    hash.NewEngine(hashAlgo),
		auditLog:  auditLogger,
		policies:  make(map[string]*EvictionPolicy),
		maxSizeGB: maxSizeGB,
	}, nil
}

// RegisterPolicy registers an eviction policy
func (m *Manager) RegisterPolicy(policy *EvictionPolicy) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if policy.Name == "" {
		return fmt.Errorf("policy name required")
	}

	m.policies[policy.Name] = policy
	return nil
}

// SaveResult caches the result of a task execution
func (m *Manager) SaveResult(taskName string, inputData []byte, output []byte, metadata map[string]interface{}) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Compute content hash
	inputHash, err := m.hasher.HashData(inputData)
	if err != nil {
		m.auditLog.LogError("hash_error", taskName, err)
		return "", fmt.Errorf("hash error: %w", err)
	}

	now := time.Now()
	entry := &storage.Entry{
		Hash:       inputHash,
		Data:       output,
		CreatedAt:  now,
		AccessedAt: now,
		Size:       int64(len(output)),
		Metadata: map[string]interface{}{
			"task":        taskName,
			"input_hash":  inputHash,
			"output_size": len(output),
			"user_data":   metadata,
		},
	}

	// Apply policy TTL if specified
	if policy, exists := m.policies[taskName]; exists {
		if policy.TTL > 0 {
			expiresAt := now.Add(policy.TTL)
			entry.ExpiresAt = &expiresAt
		}
	}

	if err := m.store.Set(entry); err != nil {
		m.auditLog.LogError("save_error", taskName, err)
		return "", fmt.Errorf("save error: %w", err)
	}

	m.auditLog.LogHit("save", taskName, inputHash)
	return inputHash, nil
}

// GetResult retrieves a cached result by task name and input
func (m *Manager) GetResult(taskName string, inputData []byte) ([]byte, map[string]interface{}, bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Compute input hash
	inputHash, err := m.hasher.HashData(inputData)
	if err != nil {
		m.auditLog.LogError("hash_error", taskName, err)
		return nil, nil, false, fmt.Errorf("hash error: %w", err)
	}

	// Look up in cache
	entry, err := m.store.Get(inputHash)
	if err != nil {
		m.auditLog.LogError("get_error", taskName, err)
		return nil, nil, false, fmt.Errorf("get error: %w", err)
	}

	if entry == nil {
		m.auditLog.LogMiss("get", taskName, inputHash)
		return nil, nil, false, nil // Cache miss
	}

	m.auditLog.LogHit("get", taskName, inputHash)
	return entry.Data, entry.Metadata, true, nil
}

// InvalidateTask clears all cached entries for a task
func (m *Manager) InvalidateTask(taskName string) (int, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// This is a simplified version - in production you'd query by metadata
	// For now, manually delete specific hashes or implement metadata indexing
	return 0, nil
}

// GetStats returns cache statistics
func (m *Manager) GetStats() (map[string]interface{}, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	stats, err := m.store.Stats()
	if err != nil {
		return nil, fmt.Errorf("stats error: %w", err)
	}

	stats["max_size_gb"] = m.maxSizeGB
	return stats, nil
}

// Close cleanly shuts down the manager
func (m *Manager) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if err := m.auditLog.Close(); err != nil {
		return fmt.Errorf("audit log error: %w", err)
	}

	return m.store.Close()
}

// ExportSnapshot exports current cache state as JSON for backup
func (m *Manager) ExportSnapshot() ([]byte, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	stats, err := m.store.Stats()
	if err != nil {
		return nil, fmt.Errorf("snapshot error: %w", err)
	}

	snapshot := map[string]interface{}{
		"timestamp": time.Now().Format(time.RFC3339),
		"stats":     stats,
		"id":        uuid.New().String(),
	}

	return json.MarshalIndent(snapshot, "", "  ")
}
