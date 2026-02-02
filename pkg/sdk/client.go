package sdk

import (
	"fmt"

	"github.com/taskvault/taskvault/internal/cache"
	"github.com/taskvault/taskvault/internal/config"
	"github.com/taskvault/taskvault/internal/hash"
)

// Client is the programmatic interface to TaskVault
type Client struct {
	manager *cache.Manager
	config  *config.Config
}

// NewClient creates a new TaskVault client
func NewClient(configPath string) (*Client, error) {
	cfg, err := config.LoadFromFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("config error: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("config validation: %w", err)
	}

	manager, err := cache.NewManager(cfg.CacheDir, cfg.MaxSizeGB, hash.HashAlgorithm(cfg.HashAlgo))
	if err != nil {
		return nil, fmt.Errorf("manager error: %w", err)
	}

	return &Client{
		manager: manager,
		config:  cfg,
	}, nil
}

// CacheResult wraps result saving with convenience
func (c *Client) CacheResult(taskName string, input []byte, output []byte) (cacheKey string, err error) {
	return c.manager.SaveResult(taskName, input, output, nil)
}

// GetCachedResult retrieves a result with hit/miss info
func (c *Client) GetCachedResult(taskName string, input []byte) (output []byte, hit bool, err error) {
	result, _, found, err := c.manager.GetResult(taskName, input)
	return result, found, err
}

// GetStats returns current cache statistics
func (c *Client) GetStats() (map[string]interface{}, error) {
	return c.manager.GetStats()
}

// Close cleanly shuts down the client
func (c *Client) Close() error {
	return c.manager.Close()
}
