package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/pelletier/go-toml/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"mneme/internal/core"
)

func TestDefaultConfigWriter(t *testing.T) {
	t.Run("returns valid TOML", func(t *testing.T) {
		configStr, err := DefaultConfigWriter()
		require.NoError(t, err)
		require.NotEmpty(t, configStr)

		// Verify it contains expected sections
		assert.Contains(t, configStr, "version = 1")
		assert.Contains(t, configStr, "[index]")
		assert.Contains(t, configStr, "[sources]")
		assert.Contains(t, configStr, "[watcher]")
		assert.Contains(t, configStr, "[search]")
		assert.Contains(t, configStr, "[ranking]")
		assert.Contains(t, configStr, "[logging]")
	})

	t.Run("contains expected default values", func(t *testing.T) {
		configStr, err := DefaultConfigWriter()
		require.NoError(t, err)

		// Check index defaults
		assert.Contains(t, configStr, "segment_size = 500")
		assert.Contains(t, configStr, "max_tokens_per_document = 10000")
		assert.Contains(t, configStr, "reindex_on_modify = true")
		assert.Contains(t, configStr, "skip_binary_files = true")

		// Check watcher defaults
		assert.Contains(t, configStr, "enabled = true")
		assert.Contains(t, configStr, "debounce_ms = 500")

		// Check search defaults
		assert.Contains(t, configStr, "default_limit = 20")
		assert.Contains(t, configStr, "use_stopwords = true")
		assert.Contains(t, configStr, "language = 'en'")

		// Check ranking defaults
		assert.Contains(t, configStr, "tfidf_weight = 1")
		assert.Contains(t, configStr, "recency_weight = 0.3")
		assert.Contains(t, configStr, "title_boost = 1.5")
		assert.Contains(t, configStr, "path_boost = 1.2")
		assert.Contains(t, configStr, "recency_half_life_days = 30")

		// Check logging defaults
		assert.Contains(t, configStr, "level = 'info'")
		assert.Contains(t, configStr, "json = true")
	})
}

func TestDefaultConfig(t *testing.T) {
	t.Run("has correct structure", func(t *testing.T) {
		config := DefaultConfig

		assert.Equal(t, uint8(1), config.Version)
		assert.NotNil(t, config.Index)
		assert.NotNil(t, config.Sources)
		assert.NotNil(t, config.Watcher)
		assert.NotNil(t, config.Search)
		assert.NotNil(t, config.Ranking)
		assert.NotNil(t, config.Logging)
	})

	t.Run("has correct index defaults", func(t *testing.T) {
		config := DefaultConfig

		assert.Equal(t, 500, config.Index.SegmentSize)
		assert.Equal(t, 10000, config.Index.MaxTokensPerDocument)
		assert.True(t, config.Index.ReindexOnModify)
		assert.True(t, config.Index.SkipBinaryFiles)
	})

	t.Run("has correct sources defaults", func(t *testing.T) {
		config := DefaultConfig

		assert.Empty(t, config.Sources.Paths)
		assert.Empty(t, config.Sources.IncludeExtensions)
		assert.Contains(t, config.Sources.Ignore, ".git")
		assert.Contains(t, config.Sources.Ignore, "node_modules")
	})

	t.Run("has correct watcher defaults", func(t *testing.T) {
		config := DefaultConfig

		assert.True(t, config.Watcher.Enabled)
		assert.Equal(t, 500, config.Watcher.DebounceMS)
	})

	t.Run("has correct search defaults", func(t *testing.T) {
		config := DefaultConfig

		assert.Equal(t, 20, config.Search.DefaultLimit)
		assert.True(t, config.Search.UseStopwords)
		assert.Equal(t, "en", config.Search.Language)
	})

	t.Run("has correct ranking defaults", func(t *testing.T) {
		config := DefaultConfig

		assert.Equal(t, 1.0, config.Ranking.TFIDFWeight)
		assert.Equal(t, 0.3, config.Ranking.RecencyWeight)
		assert.Equal(t, 1.5, config.Ranking.TitleBoost)
		assert.Equal(t, 1.2, config.Ranking.PathBoost)
		assert.Equal(t, 30, config.Ranking.RecencyHalfLifeDays)
	})

	t.Run("has correct logging defaults", func(t *testing.T) {
		config := DefaultConfig

		assert.Equal(t, "info", config.Logging.Level)
		assert.True(t, config.Logging.JSON)
	})
}

func TestConfigMarshaling(t *testing.T) {
	t.Run("can marshal and unmarshal config", func(t *testing.T) {
		// Create a test config
		testConfig := core.Config{
			Version: 1,
			Index: core.IndexConfig{
				SegmentSize:          100,
				MaxTokensPerDocument: 5000,
				ReindexOnModify:      false,
				SkipBinaryFiles:      false,
			},
			Sources: core.SourcesConfig{
				Paths:             []string{"/test/path"},
				IncludeExtensions: []string{".txt", ".md"},
				Ignore:            []string{".git"},
			},
			Watcher: core.WatcherConfig{
				Enabled:    false,
				DebounceMS: 1000,
			},
			Search: core.SearchConfig{
				DefaultLimit: 10,
				UseStopwords: false,
				Language:     "es",
			},
			Ranking: core.RankingConfig{
				TFIDFWeight:         2.0,
				RecencyWeight:       0.5,
				TitleBoost:          2.0,
				PathBoost:           1.5,
				RecencyHalfLifeDays: 60,
			},
			Logging: core.LoggingConfig{
				Level: "debug",
				JSON:  false,
			},
		}

		// Marshal to TOML
		configBytes, err := toml.Marshal(testConfig)
		require.NoError(t, err)
		require.NotEmpty(t, configBytes)

		// Unmarshal back
		var unmarshaledConfig core.Config
		err = toml.Unmarshal(configBytes, &unmarshaledConfig)
		require.NoError(t, err)

		// Verify all fields match
		assert.Equal(t, testConfig.Version, unmarshaledConfig.Version)
		assert.Equal(t, testConfig.Index, unmarshaledConfig.Index)
		assert.Equal(t, testConfig.Sources, unmarshaledConfig.Sources)
		assert.Equal(t, testConfig.Watcher, unmarshaledConfig.Watcher)
		assert.Equal(t, testConfig.Search, unmarshaledConfig.Search)
		assert.Equal(t, testConfig.Ranking, unmarshaledConfig.Ranking)
		assert.Equal(t, testConfig.Logging, unmarshaledConfig.Logging)
	})
}

func TestConfigFileOperations(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.toml")

	t.Run("write and read config file", func(t *testing.T) {
		// Create a test config
		testConfig := core.Config{
			Version: 1,
			Index: core.IndexConfig{
				SegmentSize:          500,
				MaxTokensPerDocument: 10000,
				ReindexOnModify:      true,
				SkipBinaryFiles:      true,
			},
			Sources: core.SourcesConfig{
				Paths:             []string{"/test/path1", "/test/path2"},
				IncludeExtensions: []string{".txt"},
				Ignore:            []string{".git", "node_modules"},
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
				TFIDFWeight:         1.0,
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

		// Marshal and write to file
		configBytes, err := toml.Marshal(testConfig)
		require.NoError(t, err)

		err = os.WriteFile(configPath, configBytes, 0644)
		require.NoError(t, err)

		// Read back from file
		readBytes, err := os.ReadFile(configPath)
		require.NoError(t, err)

		// Unmarshal and verify
		var readConfig core.Config
		err = toml.Unmarshal(readBytes, &readConfig)
		require.NoError(t, err)

		assert.Equal(t, testConfig, readConfig)
	})

	t.Run("handle missing config file", func(t *testing.T) {
		// Try to read a non-existent file
		_, err := os.ReadFile(filepath.Join(tempDir, "nonexistent.toml"))
		assert.Error(t, err)
		assert.True(t, os.IsNotExist(err))
	})

	t.Run("handle empty config file", func(t *testing.T) {
		// Create an empty file
		emptyPath := filepath.Join(tempDir, "empty.toml")
		err := os.WriteFile(emptyPath, []byte(""), 0644)
		require.NoError(t, err)

		// Try to unmarshal empty file
		var config core.Config
		err = toml.Unmarshal([]byte(""), &config)
		// Empty file should not error, just result in zero values
		assert.NoError(t, err)
	})
}
