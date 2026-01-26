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
var DefaultConfig = core.DefaultConfig{
	Version: 1,
	Index: core.IndexConfig{
		SegmentSize:          500,
		MaxTokensPerDocument: 10000,
		ReindexOnModify:      true,
		SkipBinaryFiles:      true,
	},
	Sources: core.SourcesConfig{
		Paths:             []string{},
		IncludeExtensions: []string{},
		Ignore:            []string{".git", "node_modules", ".vscode", ".idea", "vendor", ".cache"},
	},
	Watcher: core.WatcherConfig{
		Enabled:    true,
		DebounceMS: 500,
	},
	Search: core.SearchConfig{
		DefaultLimit: 20,
		UseStopwords: true,
		Language:     "en",
	},
	Ranking: core.RankingConfig{
		TFIDFWeight:         1,
		RecencyWeight:       0.3,
		TitleBoost:          1.5,
		PathBoost:           1.2,
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
