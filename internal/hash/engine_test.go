package hash

import (
	"testing"
)

func TestHashData(t *testing.T) {
	engine := NewEngine(Blake3)

	// Same input should produce same hash
	data1 := []byte("test input data")
	hash1, err := engine.HashData(data1)
	if err != nil {
		t.Fatalf("hash error: %v", err)
	}

	hash2, err := engine.HashData(data1)
	if err != nil {
		t.Fatalf("hash error: %v", err)
	}

	if hash1 != hash2 {
		t.Errorf("expected same hash, got %s != %s", hash1, hash2)
	}

	// Different input should produce different hash
	data2 := []byte("different input data")
	hash3, err := engine.HashData(data2)
	if err != nil {
		t.Fatalf("hash error: %v", err)
	}

	if hash1 == hash3 {
		t.Errorf("expected different hashes")
	}
}

func TestHashAlgorithmSwitching(t *testing.T) {
	data := []byte("test data")

	engineBlake3 := NewEngine(Blake3)
	hashBlake3, _ := engineBlake3.HashData(data)

	engineSHA256 := NewEngine(SHA256)
	hashSHA256, _ := engineSHA256.HashData(data)

	// Different algorithms should produce different hashes
	if hashBlake3 == hashSHA256 {
		t.Errorf("Blake3 and SHA256 should differ")
	}

	// Both should be valid hex strings
	if len(hashBlake3) != 64 && len(hashBlake3) != 128 {
		t.Errorf("unexpected hash length: %d", len(hashBlake3))
	}
}
