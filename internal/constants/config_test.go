package constants

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDirPath(t *testing.T) {
	t.Run("has expected format", func(t *testing.T) {
		// DirPath is now an absolute path from platform package
		assert.True(t, filepath.IsAbs(DirPath), "DirPath should be absolute")
		assert.Contains(t, DirPath, "mneme")
	})

	t.Run("is not empty", func(t *testing.T) {
		assert.NotEmpty(t, DirPath)
	})
}

func TestConfigPath(t *testing.T) {
	t.Run("has expected format", func(t *testing.T) {
		// ConfigPath is now an absolute path from platform package
		assert.True(t, filepath.IsAbs(ConfigPath), "ConfigPath should be absolute")
		assert.Contains(t, ConfigPath, "mneme")
		assert.True(t, strings.HasSuffix(ConfigPath, ".toml"))
	})

	t.Run("is not empty", func(t *testing.T) {
		assert.NotEmpty(t, ConfigPath)
	})
}

func TestAppName(t *testing.T) {
	t.Run("has expected value", func(t *testing.T) {
		assert.Equal(t, "mneme", AppName)
	})

	t.Run("is not empty", func(t *testing.T) {
		assert.NotEmpty(t, AppName)
	})
}
