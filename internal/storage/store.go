package storage

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// Entry represents a cached task output
type Entry struct {
	Hash       string                 `json:"hash"`
	Data       []byte                 `json:"data"`
	Metadata   map[string]interface{} `json:"metadata"`
	CreatedAt  time.Time              `json:"created_at"`
	AccessedAt time.Time              `json:"accessed_at"`
	ExpiresAt  *time.Time             `json:"expires_at,omitempty"`
	Size       int64                  `json:"size"`
}

// Store manages persistent cache storage
type Store struct {
	db        *sql.DB
	blobDir   string
	cacheSize int64 // max cache size in bytes
}

// NewStore creates/opens SQLite cache database and blob store
func NewStore(cacheDir string, maxSizeGB int64) (*Store, error) {
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return nil, fmt.Errorf("cannot create cache directory: %w", err)
	}

	dbPath := filepath.Join(cacheDir, "cache.db")
	blobDir := filepath.Join(cacheDir, "blobs")

	if err := os.MkdirAll(blobDir, 0755); err != nil {
		return nil, fmt.Errorf("cannot create blob directory: %w", err)
	}

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("cannot open database: %w", err)
	}

	// Configure SQLite for performance
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)

	store := &Store{
		db:        db,
		blobDir:   blobDir,
		cacheSize: maxSizeGB * 1024 * 1024 * 1024,
	}

	if err := store.initSchema(); err != nil {
		db.Close()
		return nil, fmt.Errorf("cannot initialize schema: %w", err)
	}

	return store, nil
}

// initSchema creates cache metadata tables
func (s *Store) initSchema() error {
	schema := `
	CREATE TABLE IF NOT EXISTS cache_entries (
		hash TEXT PRIMARY KEY,
		metadata TEXT NOT NULL,
		created_at TIMESTAMP NOT NULL,
		accessed_at TIMESTAMP NOT NULL,
		expires_at TIMESTAMP,
		size INTEGER NOT NULL,
		blob_path TEXT NOT NULL
	);

	CREATE INDEX IF NOT EXISTS idx_accessed ON cache_entries(accessed_at);
	CREATE INDEX IF NOT EXISTS idx_expires ON cache_entries(expires_at);
	CREATE INDEX IF NOT EXISTS idx_size ON cache_entries(size);
	`

	_, err := s.db.Exec(schema)
	return err
}

// Set stores a cache entry
func (s *Store) Set(entry *Entry) error {
	// Write blob to disk
	blobPath := filepath.Join(s.blobDir, entry.Hash)
	if err := os.WriteFile(blobPath, entry.Data, 0644); err != nil {
		return fmt.Errorf("cannot write blob: %w", err)
	}

	metadataJSON, err := json.Marshal(entry.Metadata)
	if err != nil {
		os.Remove(blobPath)
		return fmt.Errorf("cannot marshal metadata: %w", err)
	}

	stmt := `
	INSERT OR REPLACE INTO cache_entries 
	(hash, metadata, created_at, accessed_at, expires_at, size, blob_path)
	VALUES (?, ?, ?, ?, ?, ?, ?)
	`

	_, err = s.db.Exec(stmt,
		entry.Hash,
		string(metadataJSON),
		entry.CreatedAt,
		entry.AccessedAt,
		entry.ExpiresAt,
		entry.Size,
		blobPath,
	)

	if err != nil {
		os.Remove(blobPath)
		return fmt.Errorf("cannot insert cache entry: %w", err)
	}

	// Evict old entries if cache exceeds limit
	return s.evictIfNeeded()
}

// Get retrieves a cache entry
func (s *Store) Get(hash string) (*Entry, error) {
	stmt := `
	SELECT metadata, created_at, accessed_at, expires_at, size, blob_path
	FROM cache_entries
	WHERE hash = ? AND (expires_at IS NULL OR expires_at > datetime('now'))
	`

	var metadataJSON string
	var blobPath string
	var createdAt, accessedAt time.Time
	var expiresAt sql.NullTime
	var size int64

	err := s.db.QueryRow(stmt, hash).Scan(
		&metadataJSON, &createdAt, &accessedAt, &expiresAt, &size, &blobPath,
	)

	if err == sql.ErrNoRows {
		return nil, nil // Cache miss
	}
	if err != nil {
		return nil, fmt.Errorf("database error: %w", err)
	}

	// Read blob
	data, err := os.ReadFile(blobPath)
	if err != nil {
		// Blob missing but metadata exists - corrupted cache
		s.Delete(hash)
		return nil, nil
	}

	var metadata map[string]interface{}
	if err := json.Unmarshal([]byte(metadataJSON), &metadata); err != nil {
		return nil, fmt.Errorf("cannot unmarshal metadata: %w", err)
	}

	// Update access time
	updateStmt := `UPDATE cache_entries SET accessed_at = datetime('now') WHERE hash = ?`
	s.db.Exec(updateStmt, hash)

	expiresAtPtr := (*time.Time)(nil)
	if expiresAt.Valid {
		expiresAtPtr = &expiresAt.Time
	}

	return &Entry{
		Hash:       hash,
		Data:       data,
		Metadata:   metadata,
		CreatedAt:  createdAt,
		AccessedAt: time.Now(),
		ExpiresAt:  expiresAtPtr,
		Size:       size,
	}, nil
}

// Delete removes a cache entry
func (s *Store) Delete(hash string) error {
	stmt := `SELECT blob_path FROM cache_entries WHERE hash = ?`
	var blobPath string

	err := s.db.QueryRow(stmt, hash).Scan(&blobPath)
	if err == sql.ErrNoRows {
		return nil
	}
	if err != nil {
		return fmt.Errorf("database error: %w", err)
	}

	// Remove from database
	deleteStmt := `DELETE FROM cache_entries WHERE hash = ?`
	if _, err := s.db.Exec(deleteStmt, hash); err != nil {
		return fmt.Errorf("cannot delete entry: %w", err)
	}

	// Remove blob file
	os.Remove(blobPath)
	return nil
}

// evictIfNeeded removes least-recently-used entries if cache exceeds limit
func (s *Store) evictIfNeeded() error {
	var totalSize int64
	err := s.db.QueryRow(`SELECT COALESCE(SUM(size), 0) FROM cache_entries`).Scan(&totalSize)
	if err != nil {
		return fmt.Errorf("cannot calculate cache size: %w", err)
	}

	if totalSize <= s.cacheSize {
		return nil
	}

	// Remove oldest accessed entries until under limit
	targetSize := s.cacheSize / 2 // Keep 50% to avoid thrashing

	stmt := `
	SELECT hash, size FROM cache_entries
	ORDER BY accessed_at ASC
	`

	rows, err := s.db.Query(stmt)
	if err != nil {
		return fmt.Errorf("cannot query entries: %w", err)
	}
	defer rows.Close()

	for rows.Next() && totalSize > targetSize {
		var hash string
		var size int64

		if err := rows.Scan(&hash, &size); err != nil {
			continue
		}

		if err := s.Delete(hash); err != nil {
			continue
		}

		totalSize -= size
	}

	return nil
}

// Stats returns cache statistics
func (s *Store) Stats() (map[string]interface{}, error) {
	var count int64
	var totalSize int64
	var oldestAccess sql.NullString

	err := s.db.QueryRow(`
	SELECT COUNT(*), COALESCE(SUM(size), 0), MIN(accessed_at)
	FROM cache_entries
	`).Scan(&count, &totalSize, &oldestAccess)

	if err != nil {
		return nil, fmt.Errorf("cannot get stats: %w", err)
	}

	stats := map[string]interface{}{
		"entries":       count,
		"total_size":    totalSize,
		"cache_limit":   s.cacheSize,
		"usage_percent": float64(totalSize) / float64(s.cacheSize) * 100,
	}

	if oldestAccess.Valid {
		stats["oldest_access"] = oldestAccess.String
	}

	return stats, nil
}

// Close cleanly closes the database connection
func (s *Store) Close() error {
	return s.db.Close()
}
