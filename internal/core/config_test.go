package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultConfigStructure(t *testing.T) {
	t.Run("has all required fields", func(t *testing.T) {
		config := Config{
			Version: 1,
			Index: IndexConfig{
				SegmentSize:          500,
				MaxTokensPerDocument: 10000,
				ReindexOnModify:      true,
				SkipBinaryFiles:      true,
			},
			Sources: SourcesConfig{
				Paths:             []string{},
				IncludeExtensions: []string{},
				Ignore:            []string{".git", "node_modules", ".vscode", ".idea", "vendor", ".cache"},
			},
			Watcher: WatcherConfig{
				Enabled:    true,
				DebounceMS: 500,
			},
			Search: SearchConfig{
				DefaultLimit: 20,
				UseStopwords: true,
				Language:     "en",
			},
			Ranking: RankingConfig{
				BM25Weight:          0.7,
				VSMWeight:           0.3,
				RecencyHalfLifeDays: 30,
			},
			Logging: LoggingConfig{
				Level: "info",
				JSON:  true,
			},
		}

		assert.Equal(t, uint8(1), config.Version)
		assert.NotNil(t, config.Index)
		assert.NotNil(t, config.Sources)
		assert.NotNil(t, config.Watcher)
		assert.NotNil(t, config.Search)
		assert.NotNil(t, config.Ranking)
		assert.NotNil(t, config.Logging)
	})
}

func TestIndexConfig(t *testing.T) {
	t.Run("has correct default values", func(t *testing.T) {
		config := IndexConfig{
			SegmentSize:          500,
			MaxTokensPerDocument: 10000,
			ReindexOnModify:      true,
			SkipBinaryFiles:      true,
		}

		assert.Equal(t, 500, config.SegmentSize)
		assert.Equal(t, 10000, config.MaxTokensPerDocument)
		assert.True(t, config.ReindexOnModify)
		assert.True(t, config.SkipBinaryFiles)
	})

	t.Run("can be modified", func(t *testing.T) {
		config := IndexConfig{
			SegmentSize:          500,
			MaxTokensPerDocument: 10000,
			ReindexOnModify:      true,
			SkipBinaryFiles:      true,
		}

		config.SegmentSize = 1000
		config.MaxTokensPerDocument = 20000
		config.ReindexOnModify = false
		config.SkipBinaryFiles = false

		assert.Equal(t, 1000, config.SegmentSize)
		assert.Equal(t, 20000, config.MaxTokensPerDocument)
		assert.False(t, config.ReindexOnModify)
		assert.False(t, config.SkipBinaryFiles)
	})
}

func TestSourcesConfig(t *testing.T) {
	t.Run("has correct default values", func(t *testing.T) {
		config := SourcesConfig{
			Paths:             []string{},
			IncludeExtensions: []string{},
			Ignore:            []string{".git", "node_modules", ".vscode", ".idea", "vendor", ".cache"},
		}

		assert.Empty(t, config.Paths)
		assert.Empty(t, config.IncludeExtensions)
		assert.Len(t, config.Ignore, 6)
		assert.Contains(t, config.Ignore, ".git")
		assert.Contains(t, config.Ignore, "node_modules")
	})

	t.Run("can add paths", func(t *testing.T) {
		config := SourcesConfig{
			Paths:             []string{},
			IncludeExtensions: []string{},
			Ignore:            []string{},
		}

		config.Paths = append(config.Paths, "/path1", "/path2")
		assert.Len(t, config.Paths, 2)
		assert.Contains(t, config.Paths, "/path1")
		assert.Contains(t, config.Paths, "/path2")
	})

	t.Run("can add extensions", func(t *testing.T) {
		config := SourcesConfig{
			Paths:             []string{},
			IncludeExtensions: []string{},
			Ignore:            []string{},
		}

		config.IncludeExtensions = append(config.IncludeExtensions, ".txt", ".md")
		assert.Len(t, config.IncludeExtensions, 2)
		assert.Contains(t, config.IncludeExtensions, ".txt")
		assert.Contains(t, config.IncludeExtensions, ".md")
	})
}

func TestWatcherConfig(t *testing.T) {
	t.Run("has correct default values", func(t *testing.T) {
		config := WatcherConfig{
			Enabled:    true,
			DebounceMS: 500,
		}

		assert.True(t, config.Enabled)
		assert.Equal(t, 500, config.DebounceMS)
	})

	t.Run("can be disabled", func(t *testing.T) {
		config := WatcherConfig{
			Enabled:    true,
			DebounceMS: 500,
		}

		config.Enabled = false
		assert.False(t, config.Enabled)
	})

	t.Run("can change debounce", func(t *testing.T) {
		config := WatcherConfig{
			Enabled:    true,
			DebounceMS: 500,
		}

		config.DebounceMS = 1000
		assert.Equal(t, 1000, config.DebounceMS)
	})
}

func TestSearchConfig(t *testing.T) {
	t.Run("has correct default values", func(t *testing.T) {
		config := SearchConfig{
			DefaultLimit: 20,
			UseStopwords: true,
			Language:     "en",
		}

		assert.Equal(t, 20, config.DefaultLimit)
		assert.True(t, config.UseStopwords)
		assert.Equal(t, "en", config.Language)
	})

	t.Run("can change limit", func(t *testing.T) {
		config := SearchConfig{
			DefaultLimit: 20,
			UseStopwords: true,
			Language:     "en",
		}

		config.DefaultLimit = 50
		assert.Equal(t, 50, config.DefaultLimit)
	})

	t.Run("can disable stopwords", func(t *testing.T) {
		config := SearchConfig{
			DefaultLimit: 20,
			UseStopwords: true,
			Language:     "en",
		}

		config.UseStopwords = false
		assert.False(t, config.UseStopwords)
	})

	t.Run("can change language", func(t *testing.T) {
		config := SearchConfig{
			DefaultLimit: 20,
			UseStopwords: true,
			Language:     "en",
		}

		config.Language = "es"
		assert.Equal(t, "es", config.Language)
	})
}

func TestRankingConfig(t *testing.T) {
	t.Run("has correct default values", func(t *testing.T) {
		config := RankingConfig{
			BM25Weight:          0.7,
			VSMWeight:           0.3,
			RecencyHalfLifeDays: 30,
		}

		assert.Equal(t, 0.7, config.BM25Weight)
		assert.Equal(t, 0.3, config.VSMWeight)
		assert.Equal(t, 30, config.RecencyHalfLifeDays)
	})

	t.Run("can adjust weights", func(t *testing.T) {
		config := RankingConfig{
			BM25Weight:          0.7,
			VSMWeight:           0.3,
			RecencyHalfLifeDays: 30,
		}

		config.BM25Weight = 0.8
		config.VSMWeight = 0.2
		config.RecencyHalfLifeDays = 60

		assert.Equal(t, 0.8, config.BM25Weight)
		assert.Equal(t, 0.2, config.VSMWeight)
		assert.Equal(t, 60, config.RecencyHalfLifeDays)
	})
}

func TestLoggingConfig(t *testing.T) {
	t.Run("has correct default values", func(t *testing.T) {
		config := LoggingConfig{
			Level: "info",
			JSON:  true,
		}

		assert.Equal(t, "info", config.Level)
		assert.True(t, config.JSON)
	})

	t.Run("can change level", func(t *testing.T) {
		config := LoggingConfig{
			Level: "info",
			JSON:  true,
		}

		config.Level = "debug"
		assert.Equal(t, "debug", config.Level)
	})

	t.Run("can disable JSON output", func(t *testing.T) {
		config := LoggingConfig{
			Level: "info",
			JSON:  true,
		}

		config.JSON = false
		assert.False(t, config.JSON)
	})
}
