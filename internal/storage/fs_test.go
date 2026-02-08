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
		storageVersion, cliVersion, err := ParseVersionFile(content)
		require.NoError(t, err)
		assert.Equal(t, "1.0.0", storageVersion)
		assert.Equal(t, "1.0.0", cliVersion)
	})

	t.Run("parses version file with different versions", func(t *testing.T) {
		content := `
STORAGE_VERSION: 2.1.0
MNEME_CLI_VERSION: 3.2.1
`
		storageVersion, cliVersion, err := ParseVersionFile(content)
		require.NoError(t, err)
		assert.Equal(t, "2.1.0", storageVersion)
		assert.Equal(t, "3.2.1", cliVersion)
	})

	t.Run("returns error when STORAGE_VERSION is missing", func(t *testing.T) {
		content := `
MNEME_CLI_VERSION: 1.0.0
`
		_, _, err := ParseVersionFile(content)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "STORAGE_VERSION not found")
	})

	t.Run("handles empty content", func(t *testing.T) {
		content := ""
		_, _, err := ParseVersionFile(content)
		assert.Error(t, err)
	})

	t.Run("handles whitespace in values", func(t *testing.T) {
		content := `
STORAGE_VERSION:   1.0.0  
MNEME_CLI_VERSION:   1.0.0  
`
		storageVersion, cliVersion, err := ParseVersionFile(content)
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
	})

	t.Run("content is not empty", func(t *testing.T) {
		content := getVersionFileContents()
		assert.NotEmpty(t, content)
	})

	t.Run("can be parsed back", func(t *testing.T) {
		content := getVersionFileContents()

		storageVersion, cliVersion, err := ParseVersionFile(content)
		require.NoError(t, err)
		assert.NotEmpty(t, storageVersion)
		assert.NotEmpty(t, cliVersion)
	})
}

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		name     string
		bytes    int64
		expected string
	}{
		{"zero bytes", 0, "0 bytes"},
		{"small bytes", 512, "512 bytes"},
		{"exact KB", 1024, "1.00 KB"},
		{"KB range", 5 * 1024, "5.00 KB"},
		{"exact MB", 1024 * 1024, "1.00 MB"},
		{"MB range", 50 * 1024 * 1024, "50.00 MB"},
		{"exact GB", 1024 * 1024 * 1024, "1.00 GB"},
		{"GB range", 5 * 1024 * 1024 * 1024, "5.00 GB"},
		{"mixed GB", 1536 * 1024 * 1024, "1.50 GB"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatBytes(tt.bytes)
			if result != tt.expected {
				t.Errorf("FormatBytes(%d) = %s, expected %s", tt.bytes, result, tt.expected)
			}
		})
	}
}

func TestReadFileContents_EdgeCases(t *testing.T) {
	tmpDir := t.TempDir()

	t.Run("reads multi-line file", func(t *testing.T) {
		testPath := filepath.Join(tmpDir, "multiline.txt")
		content := "line1\nline2\nline3\n"
		err := os.WriteFile(testPath, []byte(content), 0644)
		require.NoError(t, err)

		lines, err := ReadFileContents(testPath)
		require.NoError(t, err)
		// Scanner.Scan() doesn't include the trailing empty line from the final newline
		assert.Equal(t, 3, len(lines))
		assert.Equal(t, "line1", lines[0])
		assert.Equal(t, "line3", lines[2])
	})

	t.Run("handles empty file", func(t *testing.T) {
		testPath := filepath.Join(tmpDir, "empty.txt")
		err := os.WriteFile(testPath, []byte(""), 0644)
		require.NoError(t, err)

		lines, err := ReadFileContents(testPath)
		require.NoError(t, err)
		assert.Equal(t, 0, len(lines))
	})

	t.Run("handles file with no newline at end", func(t *testing.T) {
		testPath := filepath.Join(tmpDir, "no_newline.txt")
		content := "line1\nline2"
		err := os.WriteFile(testPath, []byte(content), 0644)
		require.NoError(t, err)

		lines, err := ReadFileContents(testPath)
		require.NoError(t, err)
		assert.Equal(t, 2, len(lines))
		assert.Equal(t, "line2", lines[1])
	})

	t.Run("handles binary-looking content", func(t *testing.T) {
		testPath := filepath.Join(tmpDir, "binary.txt")
		// Write some binary-like content
		content := []byte{0x00, 0x01, 0x02, 0xFF, 0xFE}
		err := os.WriteFile(testPath, content, 0644)
		require.NoError(t, err)

		// Should not error, just read what it can
		_, err = ReadFileContents(testPath)
		// May or may not error depending on scanner behavior
		_ = err
	})
}

func TestCreateDir_EdgeCases(t *testing.T) {
	tmpDir := t.TempDir()

	t.Run("handles deeply nested paths", func(t *testing.T) {
		deepPath := filepath.Join(tmpDir, "a", "b", "c", "d", "e", "f")
		err := CreateDir(deepPath)
		require.NoError(t, err)

		info, err := os.Stat(deepPath)
		require.NoError(t, err)
		assert.True(t, info.IsDir())
	})

	t.Run("handles path with spaces", func(t *testing.T) {
		spacePath := filepath.Join(tmpDir, "path with spaces")
		err := CreateDir(spacePath)
		require.NoError(t, err)

		info, err := os.Stat(spacePath)
		require.NoError(t, err)
		assert.True(t, info.IsDir())
	})
}

func TestFileExists_EdgeCases(t *testing.T) {
	tmpDir := t.TempDir()

	t.Run("returns false for empty path", func(t *testing.T) {
		// Empty path will likely fail expansion or stat
		exists, err := FileExists("")
		// Should either error or return false
		if err == nil && exists {
			t.Error("Empty path should not exist")
		}
	})

	t.Run("handles symbolic links to files", func(t *testing.T) {
		// Create a file
		targetFile := filepath.Join(tmpDir, "target.txt")
		err := os.WriteFile(targetFile, []byte("test"), 0644)
		require.NoError(t, err)

		// Create a symlink
		linkPath := filepath.Join(tmpDir, "link.txt")
		err = os.Symlink(targetFile, linkPath)
		if err != nil {
			t.Skip("Symlink creation not supported on this system")
		}

		exists, err := FileExists(linkPath)
		require.NoError(t, err)
		assert.True(t, exists, "Symlink to file should be considered as existing file")
	})
}

func TestDirExists_EdgeCases(t *testing.T) {
	tmpDir := t.TempDir()

	t.Run("handles symbolic links to directories", func(t *testing.T) {
		// Create a directory
		targetDir := filepath.Join(tmpDir, "target_dir")
		err := os.Mkdir(targetDir, 0755)
		require.NoError(t, err)

		// Create a symlink
		linkPath := filepath.Join(tmpDir, "link_dir")
		err = os.Symlink(targetDir, linkPath)
		if err != nil {
			t.Skip("Symlink creation not supported on this system")
		}

		exists, err := DirExists(linkPath)
		require.NoError(t, err)
		assert.True(t, exists, "Symlink to directory should be considered as existing directory")
	})
}