package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config represents TaskVault configuration
type Config struct {
	CacheDir    string            `yaml:"cache_dir"`
	MaxSizeGB   int64             `yaml:"max_size_gb"`
	HashAlgo    string            `yaml:"hash_algorithm"`
	Policies    map[string]Policy `yaml:"policies"`
	LogLevel    string            `yaml:"log_level"`
	ServicePort int               `yaml:"service_port"`
}

// Policy defines eviction and caching rules per task
type Policy struct {
	TTLSeconds   int64  `yaml:"ttl_seconds"`
	MaxSizeBytes int64  `yaml:"max_size_bytes"`
	Strategy     string `yaml:"strategy"` // "lru", "lfu", "fifo"
}

// DefaultConfig returns sensible defaults
func DefaultConfig() *Config {
	return &Config{
		CacheDir:    ".taskvault/cache",
		MaxSizeGB:   10,
		HashAlgo:    "blake3",
		LogLevel:    "info",
		ServicePort: 9999,
		Policies: map[string]Policy{
			"default": {
				TTLSeconds:   86400 * 7,         // 7 days
				MaxSizeBytes: 1024 * 1024 * 100, // 100 MB
				Strategy:     "lru",
			},
		},
	}
}

// LoadFromFile loads config from YAML file
func LoadFromFile(filePath string) (*Config, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			// Return defaults if file doesn't exist
			return DefaultConfig(), nil
		}
		return nil, fmt.Errorf("cannot read config: %w", err)
	}

	cfg := DefaultConfig()
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("cannot parse config: %w", err)
	}

	return cfg, nil
}

// SaveToFile writes config to YAML file
func (c *Config) SaveToFile(filePath string) error {
	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("cannot marshal config: %w", err)
	}

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("cannot write config: %w", err)
	}

	return nil
}

// Validate checks configuration validity
func (c *Config) Validate() error {
	if c.CacheDir == "" {
		return fmt.Errorf("cache_dir is required")
	}

	if c.MaxSizeGB < 1 {
		return fmt.Errorf("max_size_gb must be >= 1")
	}

	if c.HashAlgo != "blake3" && c.HashAlgo != "sha256" {
		return fmt.Errorf("hash_algorithm must be 'blake3' or 'sha256'")
	}

	return nil
}
