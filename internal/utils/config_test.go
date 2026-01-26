package utils

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExpandFilePath(t *testing.T) {
	t.Run("expands tilde to home directory", func(t *testing.T) {
		homeDir, err := os.UserHomeDir()
		require.NoError(t, err)

		expanded, err := ExpandFilePath("~/test/path")
		require.NoError(t, err)
		assert.Equal(t, filepath.Join(homeDir, "test/path"), expanded)
	})

	t.Run("returns absolute path for absolute path", func(t *testing.T) {
		expanded, err := ExpandFilePath("/absolute/path")
		require.NoError(t, err)
		assert.Equal(t, "/absolute/path", expanded)
	})

	t.Run("converts relative path to absolute", func(t *testing.T) {
		expanded, err := ExpandFilePath("./relative/path")
		require.NoError(t, err)
		// ExpandFilePath converts relative paths to absolute
		assert.True(t, filepath.IsAbs(expanded))
		assert.True(t, strings.HasSuffix(expanded, "relative/path"))
	})

	t.Run("converts empty path to current directory", func(t *testing.T) {
		expanded, err := ExpandFilePath("")
		require.NoError(t, err)
		// Empty path becomes current directory as absolute path
		assert.True(t, filepath.IsAbs(expanded))
	})

	t.Run("handles tilde only path", func(t *testing.T) {
		homeDir, err := os.UserHomeDir()
		require.NoError(t, err)

		expanded, err := ExpandFilePath("~")
		require.NoError(t, err)
		assert.Equal(t, homeDir, expanded)
	})

	t.Run("handles tilde with slash", func(t *testing.T) {
		homeDir, err := os.UserHomeDir()
		require.NoError(t, err)

		expanded, err := ExpandFilePath("~/")
		require.NoError(t, err)
		assert.Equal(t, homeDir, expanded)
	})
}

func TestPrettyPrintConfig(t *testing.T) {
	t.Run("prints valid TOML config", func(t *testing.T) {
		tomlConfig := `
version = 1

[index]
segment_size = 500
max_tokens_per_document = 10000

[sources]
paths = ["/path1", "/path2"]
`
		// This should not error - it just prints to stdout
		err := PrettyPrintConfig([]byte(tomlConfig))
		assert.NoError(t, err)
	})

	t.Run("handles empty config", func(t *testing.T) {
		err := PrettyPrintConfig([]byte(""))
		assert.NoError(t, err)
	})

	t.Run("returns error for invalid TOML", func(t *testing.T) {
		invalidToml := `
version = 1
invalid syntax here
`
		err := PrettyPrintConfig([]byte(invalidToml))
		// Invalid TOML should return an error
		assert.Error(t, err)
	})
}

func TestPathValidation(t *testing.T) {
	t.Run("validates path with letters", func(t *testing.T) {
		path := "/path/with/letters"
		hasLetterOrNumber := false
		hasPathSeparator := false
		hasTilde := false

		for _, c := range path {
			if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') {
				hasLetterOrNumber = true
			}
			if c == '/' || c == '\\' {
				hasPathSeparator = true
			}
			if c == '~' {
				hasTilde = true
			}
		}

		assert.True(t, hasLetterOrNumber)
		assert.True(t, hasPathSeparator)
		assert.False(t, hasTilde)
	})

	t.Run("validates path with tilde", func(t *testing.T) {
		path := "~/Documents"
		hasLetterOrNumber := false
		hasPathSeparator := false
		hasTilde := false

		for _, c := range path {
			if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') {
				hasLetterOrNumber = true
			}
			if c == '/' || c == '\\' {
				hasPathSeparator = true
			}
			if c == '~' {
				hasTilde = true
			}
		}

		assert.True(t, hasLetterOrNumber)
		assert.True(t, hasPathSeparator)
		assert.True(t, hasTilde)
	})

	t.Run("invalidates path with only special characters", func(t *testing.T) {
		path := "!!!"
		hasLetterOrNumber := false
		hasPathSeparator := false
		hasTilde := false

		for _, c := range path {
			if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') {
				hasLetterOrNumber = true
			}
			if c == '/' || c == '\\' {
				hasPathSeparator = true
			}
			if c == '~' {
				hasTilde = true
			}
		}

		assert.False(t, hasLetterOrNumber)
		assert.False(t, hasPathSeparator)
		assert.False(t, hasTilde)
	})
}
