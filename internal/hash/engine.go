package hash

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/zeebo/blake3"
)

// HashAlgorithm defines the hashing algorithm to use
type HashAlgorithm string

const (
	Blake3 HashAlgorithm = "blake3"
	SHA256 HashAlgorithm = "sha256"
)

// Engine computes content-aware hashes for arbitrary data
type Engine struct {
	algorithm HashAlgorithm
}

// NewEngine creates a new hash engine with specified algorithm
func NewEngine(algo HashAlgorithm) *Engine {
	if algo == "" {
		algo = Blake3 // default to Blake3 for performance
	}
	return &Engine{algorithm: algo}
}

// HashData computes hash of raw byte data
func (e *Engine) HashData(data []byte) (string, error) {
	switch e.algorithm {
	case Blake3:
		return e.hashBlake3(data), nil
	case SHA256:
		return e.hashSHA256(data), nil
	default:
		return "", fmt.Errorf("unsupported algorithm: %s", e.algorithm)
	}
}

// HashFile computes hash of a file's contents
func (e *Engine) HashFile(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("cannot open file %s: %w", filePath, err)
	}
	defer file.Close()

	switch e.algorithm {
	case Blake3:
		h := blake3.New()
		if _, err := io.Copy(h, file); err != nil {
			return "", fmt.Errorf("hash error for %s: %w", filePath, err)
		}
		return hex.EncodeToString(h.Sum(nil)), nil

	case SHA256:
		h := sha256.New()
		if _, err := io.Copy(h, file); err != nil {
			return "", fmt.Errorf("hash error for %s: %w", filePath, err)
		}
		return hex.EncodeToString(h.Sum(nil)), nil

	default:
		return "", fmt.Errorf("unsupported algorithm: %s", e.algorithm)
	}
}

// HashDirectory computes a composite hash of a directory tree
// Files are sorted for consistency across different file system orderings
func (e *Engine) HashDirectory(dirPath string) (string, error) {
	h := blake3.New()

	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Include relative path in hash (affects structure changes)
		relPath, err := filepath.Rel(dirPath, path)
		if err != nil {
			return err
		}

		if info.IsDir() {
			fmt.Fprintf(h, "dir:%s\n", relPath)
		} else {
			// Include file size and name in hash
			fmt.Fprintf(h, "file:%s:%d\n", relPath, info.Size())

			// Include file contents
			file, err := os.Open(path)
			if err != nil {
				return fmt.Errorf("cannot read file %s: %w", path, err)
			}
			defer file.Close()

			if _, err := io.Copy(h, file); err != nil {
				return fmt.Errorf("hash error for %s: %w", path, err)
			}
		}
		return nil
	})

	if err != nil {
		return "", fmt.Errorf("directory hash error: %w", err)
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}

// hashBlake3 computes Blake3 hash (fastest for large files)
func (e *Engine) hashBlake3(data []byte) string {
	hash := blake3.Sum256(data)
	return hex.EncodeToString(hash[:])
}

// hashSHA256 computes SHA256 hash (standard, widely supported)
func (e *Engine) hashSHA256(data []byte) string {
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}
