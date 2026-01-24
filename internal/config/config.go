package config

import (
	"fmt"

	"github.com/pelletier/go-toml/v2"
	"github.com/spf13/cobra"

	"mneme/internal/core"
	"mneme/internal/logger"
)

var ConfigCmd = &cobra.Command{
	Use:   "config",
	Short: "Configuration commands",
	Long:  `Configuration commands`,
}

var showCmd = &cobra.Command{
	Use:   "show",
	Short: "show configuration values",
	Long:  `show configuration values`,
	Run:   showCmdExecute,
}

func init() {
	ConfigCmd.AddCommand(showCmd)
}

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
