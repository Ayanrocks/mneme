package display

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFormatSearchResult_Regex(t *testing.T) {
	// Create a temp file
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.txt")
	content := "hello world\nregex test match\nanother line"
	err := os.WriteFile(tmpFile, []byte(content), 0644)
	require.NoError(t, err)

	// Since storage.ReadFileContents uses utils.ExpandFilePath, 
	// we need to make sure the file is accessible.
	// For testing, I'll rely on the fact that ReadFileContents just opens the path provided.
	
	t.Run("matches with regex", func(t *testing.T) {
		res, err := FormatSearchResult(tmpFile, nil, "regex.*match", 1.0)
		require.NoError(t, err)
		assert.NotEmpty(t, res.Snippets)
		assert.Equal(t, "regex test match", res.Snippets[0].Content)
	})

	t.Run("no match with invalid regex", func(t *testing.T) {
		res, err := FormatSearchResult(tmpFile, nil, "nonexistent", 1.0)
		require.NoError(t, err)
		assert.Empty(t, res.Snippets)
	})
}
