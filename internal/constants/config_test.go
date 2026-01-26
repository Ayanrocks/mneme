package constants

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDirPath(t *testing.T) {
	t.Run("has expected format", func(t *testing.T) {
		assert.True(t, strings.HasPrefix(DirPath, "~"))
		assert.Contains(t, DirPath, "mneme")
	})

	t.Run("is not empty", func(t *testing.T) {
		assert.NotEmpty(t, DirPath)
	})
}

func TestConfigPath(t *testing.T) {
	t.Run("has expected format", func(t *testing.T) {
		assert.True(t, strings.HasPrefix(ConfigPath, "~"))
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
