package storage

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateDir(t *testing.T) {
	t.Run("creates directory successfully", func(t *testing.T) {
		tempDir := t.TempDir()
		testPath := filepath.Join(tempDir, "test_dir")

		err := CreateDir(testPath)
		require.NoError(t, err)

		// Verify directory was created
		info, err := os.Stat(testPath)
		require.NoError(t, err)
		assert.True(t, info.IsDir())
	})

	t.Run("creates nested directories", func(t *testing.T) {
		tempDir := t.TempDir()
		testPath := filepath.Join(tempDir, "nested", "dir", "structure")

		err := CreateDir(testPath)
		require.NoError(t, err)

		// Verify directory was created
		info, err := os.Stat(testPath)
		require.NoError(t, err)
		assert.True(t, info.IsDir())
	})

	t.Run("does not error on existing directory", func(t *testing.T) {
		tempDir := t.TempDir()
		testPath := filepath.Join(tempDir, "existing")

		// Create the directory first
		err := os.Mkdir(testPath, os.ModePerm)
		require.NoError(t, err)

		// CreateDir should not error
		err = CreateDir(testPath)
		assert.NoError(t, err)
	})
}

func TestCreateFile(t *testing.T) {
	t.Run("creates file successfully", func(t *testing.T) {
		tempDir := t.TempDir()
		testPath := filepath.Join(tempDir, "test_file.txt")

		file, err := CreateFile(testPath)
		require.NoError(t, err)
		require.NotNil(t, file)
		defer file.Close()

		// Verify file was created
		info, err := os.Stat(testPath)
		require.NoError(t, err)
		assert.False(t, info.IsDir())
	})

	t.Run("creates file and allows writing", func(t *testing.T) {
		tempDir := t.TempDir()
		testPath := filepath.Join(tempDir, "writable_file.txt")

		file, err := CreateFile(testPath)
		require.NoError(t, err)
		require.NotNil(t, file)

		// Write to the file
		_, err = file.WriteString("test content")
		require.NoError(t, err)
		file.Close()

		// Verify content
		content, err := os.ReadFile(testPath)
		require.NoError(t, err)
		assert.Equal(t, "test content", string(content))
	})
}

func TestFileExists(t *testing.T) {
	t.Run("returns true for existing file", func(t *testing.T) {
		tempDir := t.TempDir()
		testPath := filepath.Join(tempDir, "existing_file.txt")

		// Create the file
		f, err := os.Create(testPath)
		require.NoError(t, err)
		f.Close()

		exists, err := FileExists(testPath)
		require.NoError(t, err)
		assert.True(t, exists)
	})

	t.Run("returns false for non-existing file", func(t *testing.T) {
		tempDir := t.TempDir()
		testPath := filepath.Join(tempDir, "non_existing_file.txt")

		exists, err := FileExists(testPath)
		require.NoError(t, err)
		assert.False(t, exists)
	})

	t.Run("returns false for directory", func(t *testing.T) {
		tempDir := t.TempDir()
		testPath := filepath.Join(tempDir, "a_directory")

		// Create a directory
		err := os.Mkdir(testPath, os.ModePerm)
		require.NoError(t, err)

		exists, err := FileExists(testPath)
		require.NoError(t, err)
		assert.False(t, exists, "FileExists should return false for directories")
	})
}

func TestDirExists(t *testing.T) {
	t.Run("returns true for existing directory", func(t *testing.T) {
		tempDir := t.TempDir()
		testPath := filepath.Join(tempDir, "existing_dir")

		// Create the directory
		err := os.Mkdir(testPath, os.ModePerm)
		require.NoError(t, err)

		exists, err := DirExists(testPath)
		require.NoError(t, err)
		assert.True(t, exists)
	})

	t.Run("returns false for non-existing directory", func(t *testing.T) {
		tempDir := t.TempDir()
		testPath := filepath.Join(tempDir, "non_existing_dir")

		exists, err := DirExists(testPath)
		require.NoError(t, err)
		assert.False(t, exists)
	})

	t.Run("returns false for file", func(t *testing.T) {
		tempDir := t.TempDir()
		testPath := filepath.Join(tempDir, "a_file.txt")

		// Create a file
		f, err := os.Create(testPath)
		require.NoError(t, err)
		f.Close()

		exists, err := DirExists(testPath)
		require.NoError(t, err)
		assert.False(t, exists, "DirExists should return false for files")
	})
}

func TestParseVersionFile(t *testing.T) {
	t.Run("parses valid version file", func(t *testing.T) {
		content := `
STORAGE_VERSION: 1.0.0
MNEME_CLI_VERSION: 1.0.0
`
		storageVersion, cliVersion, platformStr, err := ParseVersionFile(content)
		require.NoError(t, err)
		assert.Equal(t, "1.0.0", storageVersion)
		assert.Equal(t, "1.0.0", cliVersion)
		assert.Empty(t, platformStr) // No platform in this content
	})

	t.Run("parses version file with platform", func(t *testing.T) {
		content := `
STORAGE_VERSION: 1.0.0
MNEME_CLI_VERSION: 1.0.0
PLATFORM: linux
`
		storageVersion, cliVersion, platformStr, err := ParseVersionFile(content)
		require.NoError(t, err)
		assert.Equal(t, "1.0.0", storageVersion)
		assert.Equal(t, "1.0.0", cliVersion)
		assert.Equal(t, "linux", platformStr)
	})

	t.Run("parses version file with different versions", func(t *testing.T) {
		content := `
STORAGE_VERSION: 2.1.0
MNEME_CLI_VERSION: 3.2.1
`
		storageVersion, cliVersion, _, err := ParseVersionFile(content)
		require.NoError(t, err)
		assert.Equal(t, "2.1.0", storageVersion)
		assert.Equal(t, "3.2.1", cliVersion)
	})

	t.Run("returns error when STORAGE_VERSION is missing", func(t *testing.T) {
		content := `
MNEME_CLI_VERSION: 1.0.0
`
		_, _, _, err := ParseVersionFile(content)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "STORAGE_VERSION not found")
	})

	t.Run("handles empty content", func(t *testing.T) {
		content := ""
		_, _, _, err := ParseVersionFile(content)
		assert.Error(t, err)
	})

	t.Run("handles whitespace in values", func(t *testing.T) {
		content := `
STORAGE_VERSION:   1.0.0  
MNEME_CLI_VERSION:   1.0.0  
`
		storageVersion, cliVersion, _, err := ParseVersionFile(content)
		require.NoError(t, err)
		assert.Equal(t, "1.0.0", storageVersion)
		assert.Equal(t, "1.0.0", cliVersion)
	})
}

func TestGetVersionFileContents(t *testing.T) {
	t.Run("returns formatted version content", func(t *testing.T) {
		content := getVersionFileContents()

		assert.Contains(t, content, "STORAGE_VERSION:")
		assert.Contains(t, content, "MNEME_CLI_VERSION:")
		assert.Contains(t, content, "PLATFORM:")
	})

	t.Run("content is not empty", func(t *testing.T) {
		content := getVersionFileContents()
		assert.NotEmpty(t, content)
	})

	t.Run("can be parsed back", func(t *testing.T) {
		content := getVersionFileContents()

		storageVersion, cliVersion, platformStr, err := ParseVersionFile(content)
		require.NoError(t, err)
		assert.NotEmpty(t, storageVersion)
		assert.NotEmpty(t, cliVersion)
		assert.NotEmpty(t, platformStr)
	})
}
