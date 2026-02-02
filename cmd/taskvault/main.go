package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/taskvault/taskvault/internal/cache"
	"github.com/taskvault/taskvault/internal/config"
	"github.com/taskvault/taskvault/internal/hash"
)

var (
	version = "0.1.0"
	cfgFile string
	verbose bool
)

var rootCmd = &cobra.Command{
	Use:     "taskvault",
	Short:   "Intelligent task result caching for CI/CD and data pipelines",
	Long:    `TaskVault: Cache results of deterministic tasks using content-aware hashing. Never recompute the same work.`,
	Version: version,
}

var cacheCmd = &cobra.Command{
	Use:   "cache",
	Short: "Cache operations",
}

var saveCmd = &cobra.Command{
	Use:   "save <task_name> <input_file> <output_file>",
	Short: "Save task output to cache",
	Args:  cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		taskName := args[0]
		inputFile := args[1]
		outputFile := args[2]

		cfg, err := config.LoadFromFile(cfgFile)
		if err != nil {
			return err
		}

		if err := cfg.Validate(); err != nil {
			return err
		}

		manager, err := cache.NewManager(cfg.CacheDir, cfg.MaxSizeGB, hash.HashAlgorithm(cfg.HashAlgo))
		if err != nil {
			return err
		}
		defer manager.Close()

		// Read input and output files
		inputData, err := os.ReadFile(inputFile)
		if err != nil {
			return fmt.Errorf("cannot read input: %w", err)
		}

		outputData, err := os.ReadFile(outputFile)
		if err != nil {
			return fmt.Errorf("cannot read output: %w", err)
		}

		// Save to cache
		hash, err := manager.SaveResult(taskName, inputData, outputData, nil)
		if err != nil {
			return err
		}

		fmt.Printf("✓ Cached %s (hash: %s, size: %d bytes)\n", taskName, hash[:12], len(outputData))
		return nil
	},
}

var getCmd = &cobra.Command{
	Use:   "get <task_name> <input_file> <output_file>",
	Short: "Retrieve cached result or indicate miss",
	Args:  cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		taskName := args[0]
		inputFile := args[1]
		outputFile := args[2]

		cfg, err := config.LoadFromFile(cfgFile)
		if err != nil {
			return err
		}

		if err := cfg.Validate(); err != nil {
			return err
		}

		manager, err := cache.NewManager(cfg.CacheDir, cfg.MaxSizeGB, hash.HashAlgorithm(cfg.HashAlgo))
		if err != nil {
			return err
		}
		defer manager.Close()

		// Read input file
		inputData, err := os.ReadFile(inputFile)
		if err != nil {
			return fmt.Errorf("cannot read input: %w", err)
		}

		// Get from cache
		output, metadata, hit, err := manager.GetResult(taskName, inputData)
		if err != nil {
			return err
		}

		if !hit {
			fmt.Printf("✗ Cache miss for %s\n", taskName)
			return nil
		}

		// Write output file
		if err := os.WriteFile(outputFile, output, 0644); err != nil {
			return fmt.Errorf("cannot write output: %w", err)
		}

		fmt.Printf("✓ Cache hit for %s (size: %d bytes)\n", taskName, len(output))
		if verbose {
			fmt.Printf("  Metadata: %+v\n", metadata)
		}

		return nil
	},
}

var statsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Show cache statistics",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadFromFile(cfgFile)
		if err != nil {
			return err
		}

		if err := cfg.Validate(); err != nil {
			return err
		}

		manager, err := cache.NewManager(cfg.CacheDir, cfg.MaxSizeGB, hash.HashAlgorithm(cfg.HashAlgo))
		if err != nil {
			return err
		}
		defer manager.Close()

		stats, err := manager.GetStats()
		if err != nil {
			return err
		}

		fmt.Printf("TaskVault Cache Statistics\n")
		fmt.Printf("==========================\n")
		fmt.Printf("Entries:        %v\n", stats["entries"])
		fmt.Printf("Total Size:     %.2f MB\n", float64(stats["total_size"].(int64))/1024/1024)
		fmt.Printf("Cache Limit:    %.2f GB\n", float64(stats["cache_limit"].(int64))/1024/1024/1024)
		fmt.Printf("Usage:          %.1f%%\n", stats["usage_percent"])

		return nil
	},
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new configuration",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := config.DefaultConfig()

		cfgPath := filepath.Join(".taskvault", "config.yaml")
		if err := os.MkdirAll(filepath.Dir(cfgPath), 0755); err != nil {
			return fmt.Errorf("cannot create directory: %w", err)
		}

		if err := cfg.SaveToFile(cfgPath); err != nil {
			return err
		}

		fmt.Printf("✓ Created config at %s\n", cfgPath)
		return nil
	},
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", ".taskvault/config.yaml", "config file path")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")

	rootCmd.AddCommand(initCmd)

	cacheCmd.AddCommand(saveCmd, getCmd, statsCmd)
	rootCmd.AddCommand(cacheCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
