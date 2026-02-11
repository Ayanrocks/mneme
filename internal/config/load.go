package config

import (
	"fmt"
	"mneme/internal/core"
	"os"

	"github.com/pelletier/go-toml/v2"
	"github.com/spf13/cobra"

	"mneme/internal/constants"
	"mneme/internal/logger"
	"mneme/internal/utils"
)

// DefaultConfig is the default configuration for mneme
var DefaultConfig = core.Config{
	Version: 1,
	Index: core.IndexConfig{
		SegmentSize:          500,
		MaxTokensPerDocument: 0, // 0 disables the per-document token limit
		ReindexOnModify:      true,
		SkipBinaryFiles:      true,
	},
	Sources: core.SourcesConfig{
		Paths:             []string{},
		IncludeExtensions: []string{},
		ExcludeExtensions: []string{},
		Ignore:            []string{".git", "node_modules", ".vscode", ".idea", "vendor", ".cache", "target", "build"},
		Filesystem: core.FilesystemSourceConfig{
			Enabled: true,
		},
	},
	// Watcher: core.WatcherConfig{
	// 	Enabled:    true,
	// 	DebounceMS: 500,
	// },
	Search: core.SearchConfig{
		DefaultLimit: 20,
		UseStopwords: true,
		Language:     "en",
	},
	Ranking: core.RankingConfig{
		BM25Weight:          0.7,
		VSMWeight:           0.3,
		RecencyHalfLifeDays: 30,
	},
	Logging: core.LoggingConfig{
		Level: "info",
		JSON:  true,
	},
}

// DefaultConfigWriter returns the default configuration as a TOML string
func DefaultConfigWriter() (string, error) {
	configBytes, err := toml.Marshal(DefaultConfig)
	if err != nil {
		logger.Errorf("Error marshaling default config to TOML: %+v", err)
		return "", fmt.Errorf("failed to marshal default config: %w", err)
	}

	return string(configBytes), nil
}

func LoadConfig() (*core.Config, error) {
	var config core.Config

	// read config from config path
	configPath, err := utils.ExpandFilePath(constants.ConfigPath)
	if err != nil {
		logger.Errorf("Failed to expand config path: %+v", err)
		return nil, fmt.Errorf("failed to expand config path: %w", err)
	}

	// Read the config file
	configBytes, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			logger.Warnf("Config file not found at: %s", configPath)
			logger.Info("Using default configuration")
			configStr, err := DefaultConfigWriter()
			if err != nil {
				logger.Errorf("Failed to write default config: %+v", err)
				return nil, fmt.Errorf("failed to write default config: %w", err)
			}
			configBytes = []byte(configStr)
		} else {
			logger.Errorf("Failed to read config file: %+v", err)
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	if err := toml.Unmarshal(configBytes, &config); err != nil {
		logger.Errorf("Error unmarshaling config: %+v", err)
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}
	return &config, nil
}

func ShowCmdExecute(cmd *cobra.Command, args []string) {
	logger.Info("showCmdExecute")

	// Expand the config path (handle ~)
	configPath, err := utils.ExpandFilePath(constants.ConfigPath)
	if err != nil {
		logger.Errorf("Failed to expand config path: %+v", err)
		return
	}

	// Read the config file
	configBytes, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			logger.Warnf("Config file not found at: %s", configPath)
			logger.Info("Using default configuration")
			configStr, err := DefaultConfigWriter()
			if err != nil {
				logger.Errorf("Failed to write default config: %+v", err)
				return
			}
			configBytes = []byte(configStr)
		} else {
			logger.Errorf("Failed to read config file: %+v", err)
			return
		}
	}

	// Pretty print the config
	if err := utils.PrettyPrintConfig(configBytes); err != nil {
		logger.Errorf("Failed to pretty print config: %+v", err)
	}
}
