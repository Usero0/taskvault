package audit

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Logger tracks cache operations for auditing and analytics
type Logger struct {
	filePath string
	file     *os.File
	mu       sync.Mutex
}

// NewLogger creates an audit log file in the cache directory
func NewLogger(cacheDir string) (*Logger, error) {
	filePath := filepath.Join(cacheDir, "audit.log")

	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("cannot open audit log: %w", err)
	}

	return &Logger{
		filePath: filePath,
		file:     file,
	}, nil
}

// LogHit records a cache hit
func (l *Logger) LogHit(operation, taskName, hash string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	entry := fmt.Sprintf("[%s] HIT %s task=%s hash=%s\n",
		time.Now().Format(time.RFC3339),
		operation,
		taskName,
		truncateHash(hash),
	)

	l.file.WriteString(entry)
}

// LogMiss records a cache miss
func (l *Logger) LogMiss(operation, taskName, hash string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	entry := fmt.Sprintf("[%s] MISS %s task=%s hash=%s\n",
		time.Now().Format(time.RFC3339),
		operation,
		taskName,
		truncateHash(hash),
	)

	l.file.WriteString(entry)
}

// LogError records an error
func (l *Logger) LogError(errorType, taskName string, err error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	entry := fmt.Sprintf("[%s] ERROR %s task=%s error=%v\n",
		time.Now().Format(time.RFC3339),
		errorType,
		taskName,
		err,
	)

	l.file.WriteString(entry)
}

// truncateHash returns first 12 chars of hash for readability
func truncateHash(h string) string {
	if len(h) > 12 {
		return h[:12] + "..."
	}
	return h
}

// Close cleanly closes the audit log
func (l *Logger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.file != nil {
		return l.file.Close()
	}
	return nil
}
