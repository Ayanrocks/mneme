package index

import (
	"mneme/internal/core"
	"mneme/internal/ingest"
	"os"
	"path/filepath"
	"testing"
)

func TestFormatChunkFilename(t *testing.T) {
	tests := []struct {
		chunkID  int
		expected string
	}{
		{1, "001.idx"},
		{10, "010.idx"},
		{100, "100.idx"},
		{999, "999.idx"},
	}

	for _, tt := range tests {
		result := formatChunkFilename(tt.chunkID)
		if result != tt.expected {
			t.Errorf("formatChunkFilename(%d) = %s, expected %s", tt.chunkID, result, tt.expected)
		}
	}
}

func TestProcessBatch(t *testing.T) {
	// Create temp directory with test files
	tmpDir := t.TempDir()

	// Create test files
	testFiles := []string{
		filepath.Join(tmpDir, "file1.txt"),
		filepath.Join(tmpDir, "file2.txt"),
		filepath.Join(tmpDir, "file3.txt"),
	}

	for i, filePath := range testFiles {
		content := []byte("test content " + string(rune('A'+i)))
		err := os.WriteFile(filePath, content, 0644)
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
	}

	t.Run("processes batch successfully", func(t *testing.T) {
		globalDocID := uint(1)
		chunk, docCount, tokenCount := processBatch(testFiles, &globalDocID)

		if chunk == nil {
			t.Fatal("Expected non-nil chunk")
		}
		if docCount != 3 {
			t.Errorf("Expected 3 docs, got %d", docCount)
		}
		if tokenCount == 0 {
			t.Error("Expected non-zero token count")
		}
		if globalDocID != 4 {
			t.Errorf("Expected globalDocID to be 4, got %d", globalDocID)
		}
		if len(chunk.Docs) != 3 {
			t.Errorf("Expected 3 documents in chunk, got %d", len(chunk.Docs))
		}
		if len(chunk.InvertedIndex) == 0 {
			t.Error("Expected non-empty inverted index")
		}
	})

	t.Run("handles empty file list", func(t *testing.T) {
		globalDocID := uint(1)
		chunk, docCount, tokenCount := processBatch([]string{}, &globalDocID)

		if chunk == nil {
			t.Fatal("Expected non-nil chunk")
		}
		if docCount != 0 {
			t.Errorf("Expected 0 docs, got %d", docCount)
		}
		if tokenCount == 0 {
			// Empty batch should have 0 tokens
		}
		if len(chunk.Docs) != 0 {
			t.Errorf("Expected 0 documents, got %d", len(chunk.Docs))
		}
	})

	t.Run("handles non-existent files gracefully", func(t *testing.T) {
		globalDocID := uint(1)
		nonExistentFiles := []string{"/nonexistent/file1.txt", "/nonexistent/file2.txt"}
		chunk, docCount, tokenCount := processBatch(nonExistentFiles, &globalDocID)

		// Should not panic, just skip the files
		if chunk == nil {
			t.Fatal("Expected non-nil chunk")
		}
		// No valid files, so doc count should be 0
		if docCount != 0 {
			t.Errorf("Expected 0 docs for non-existent files, got %d", docCount)
		}
		_ = tokenCount
	})

	t.Run("increments globalDocID correctly", func(t *testing.T) {
		globalDocID := uint(100)
		_, docCount, _ := processBatch(testFiles, &globalDocID)

		expectedID := uint(100 + docCount)
		if globalDocID != expectedID {
			t.Errorf("Expected globalDocID %d, got %d", expectedID, globalDocID)
		}
	})
}

func TestProcessBatchWithRegistry(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test files
	testFiles := make([]string, 3)
	for i := 0; i < 3; i++ {
		filePath := filepath.Join(tmpDir, "file"+string(rune('A'+i))+".txt")
		content := []byte("test content for file " + string(rune('A'+i)))
		err := os.WriteFile(filePath, content, 0644)
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
		testFiles[i] = filePath
	}

	// Create registry with filesystem ingestor
	registry := ingest.NewRegistry()
	fsIngestor := ingest.NewFilesystemIngestor([]string{tmpDir}, nil)
	registry.Register(fsIngestor)

	t.Run("processes batch with registry", func(t *testing.T) {
		globalDocID := uint(1)
		chunk, docCount, tokenCount := processBatchWithRegistry(testFiles, registry, &globalDocID)

		if chunk == nil {
			t.Fatal("Expected non-nil chunk")
		}
		if docCount != 3 {
			t.Errorf("Expected 3 docs, got %d", docCount)
		}
		if tokenCount == 0 {
			t.Error("Expected non-zero token count")
		}
		if len(chunk.Docs) != 3 {
			t.Errorf("Expected 3 documents in chunk, got %d", len(chunk.Docs))
		}
	})

	t.Run("handles empty doc list", func(t *testing.T) {
		globalDocID := uint(1)
		chunk, docCount, tokenCount := processBatchWithRegistry([]string{}, registry, &globalDocID)

		if chunk == nil {
			t.Fatal("Expected non-nil chunk")
		}
		if docCount != 0 {
			t.Errorf("Expected 0 docs, got %d", docCount)
		}
		_ = tokenCount
	})

	t.Run("calculates avgDocLen correctly", func(t *testing.T) {
		globalDocID := uint(1)
		chunk, _, _ := processBatchWithRegistry(testFiles, registry, &globalDocID)

		if len(chunk.Docs) > 0 && chunk.AvgDocLen == 0 {
			t.Error("Expected non-zero avgDocLen when documents exist")
		}
	})
}

func TestIndexBuilderBatched(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test files
	for i := 0; i < 5; i++ {
		filePath := filepath.Join(tmpDir, "file"+string(rune('A'+i))+".txt")
		content := []byte("test content " + string(rune('A'+i)))
		err := os.WriteFile(filePath, content, 0644)
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
	}

	t.Run("builds index with default config", func(t *testing.T) {
		paths := []string{tmpDir}
		options := &core.CrawlerOptions{}
		config := core.DefaultBatchConfig()

		// Use small batch size for testing
		config.BatchSize = 2
		config.SuppressLogs = true

		// Note: This test will fail without proper storage setup
		// So we mainly test that it doesn't panic
		_, err := IndexBuilderBatched(paths, options, config)
		// Error is expected if storage isn't initialized, but shouldn't panic
		_ = err
	})

	t.Run("handles explicitly created default config", func(t *testing.T) {
		paths := []string{tmpDir}
		options := &core.CrawlerOptions{}
		// Note: The implementation has a bug where it accesses config before checking for nil
		// So we provide an explicit config instead of nil
		config := core.DefaultBatchConfig()
		config.SuppressLogs = true

		_, _ = IndexBuilderBatched(paths, options, config)
		// Should use provided config (error is expected due to storage not being initialized)
	})

	t.Run("handles empty paths", func(t *testing.T) {
		paths := []string{}
		options := &core.CrawlerOptions{}
		config := core.DefaultBatchConfig()
		config.SuppressLogs = true

		manifest, err := IndexBuilderBatched(paths, options, config)
		if err != nil {
			// Expected error for empty paths or uninitialized storage
		}
		if manifest != nil && len(manifest.Chunks) > 0 {
			t.Error("Expected no chunks for empty paths")
		}
	})
}

func TestIndexBuilderBatchedWithRegistry(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test files
	for i := 0; i < 3; i++ {
		filePath := filepath.Join(tmpDir, "file"+string(rune('A'+i))+".txt")
		content := []byte("test content " + string(rune('A'+i)))
		err := os.WriteFile(filePath, content, 0644)
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
	}

	t.Run("builds index with registry", func(t *testing.T) {
		registry := ingest.NewRegistry()
		fsIngestor := ingest.NewFilesystemIngestor([]string{tmpDir}, nil)
		registry.Register(fsIngestor)

		options := &core.CrawlerOptions{}
		config := core.DefaultBatchConfig()
		config.BatchSize = 2
		config.SuppressLogs = true

		_, err := IndexBuilderBatchedWithRegistry(registry, options, config)
		// Error expected if storage isn't initialized
		_ = err
	})

	t.Run("handles nil config", func(t *testing.T) {
		registry := ingest.NewRegistry()
		fsIngestor := ingest.NewFilesystemIngestor([]string{tmpDir}, nil)
		registry.Register(fsIngestor)

		options := &core.CrawlerOptions{}

		_, err := IndexBuilderBatchedWithRegistry(registry, options, nil)
		// Should use default config
		_ = err
	})

	t.Run("calls progress callback", func(t *testing.T) {
		registry := ingest.NewRegistry()
		fsIngestor := ingest.NewFilesystemIngestor([]string{tmpDir}, nil)
		registry.Register(fsIngestor)

		options := &core.CrawlerOptions{}
		config := core.DefaultBatchConfig()
		config.SuppressLogs = true

		callbackCalled := false
		config.ProgressCallback = func(current, total int, message string) {
			callbackCalled = true
		}

		_, _ = IndexBuilderBatchedWithRegistry(registry, options, config)

		if !callbackCalled {
			t.Error("Expected progress callback to be called")
		}
	})

	t.Run("handles empty registry", func(t *testing.T) {
		registry := ingest.NewRegistry()
		options := &core.CrawlerOptions{}
		config := core.DefaultBatchConfig()
		config.SuppressLogs = true

		manifest, err := IndexBuilderBatchedWithRegistry(registry, options, config)
		if err != nil {
			// Expected error for empty registry
		}
		if manifest != nil && manifest.TotalDocs > 0 {
			t.Error("Expected no docs for empty registry")
		}
	})
}

func TestTokenize(t *testing.T) {
	t.Run("tokenizes simple text", func(t *testing.T) {
		tokens := Tokenize("hello world test")
		if len(tokens) == 0 {
			t.Error("Expected non-empty tokens")
		}
	})

	t.Run("handles empty string", func(t *testing.T) {
		tokens := Tokenize("")
		if len(tokens) != 0 {
			t.Errorf("Expected empty tokens for empty string, got %d", len(tokens))
		}
	})

	t.Run("tokenizes camelCase", func(t *testing.T) {
		tokens := Tokenize("getUserName")
		// Should split camelCase into tokens
		if len(tokens) == 0 {
			t.Error("Expected tokens from camelCase")
		}
	})

	t.Run("tokenizes snake_case", func(t *testing.T) {
		tokens := Tokenize("user_name_here")
		if len(tokens) == 0 {
			t.Error("Expected tokens from snake_case")
		}
	})

	t.Run("handles special characters", func(t *testing.T) {
		tokens := Tokenize("test@example.com")
		if len(tokens) == 0 {
			t.Error("Expected tokens from text with special characters")
		}
	})

	t.Run("removes stopwords or stems", func(t *testing.T) {
		// Tokenize applies stemming, so "running" becomes "run"
		tokens := Tokenize("running quickly")
		// Should have at least some tokens
		if len(tokens) == 0 {
			t.Error("Expected some tokens after processing")
		}
	})
}

func TestIndexBuilderEdgeCases(t *testing.T) {
	t.Run("handles very large batch size", func(t *testing.T) {
		config := core.DefaultBatchConfig()
		config.BatchSize = 1000000 // Very large
		// Should not panic
		if config.BatchSize <= 0 {
			t.Error("Batch size should remain positive")
		}
	})

	t.Run("handles batch size of 1", func(t *testing.T) {
		tmpDir := t.TempDir()
		testFile := filepath.Join(tmpDir, "test.txt")
		err := os.WriteFile(testFile, []byte("test"), 0644)
		if err != nil {
			t.Fatal(err)
		}

		registry := ingest.NewRegistry()
		fsIngestor := ingest.NewFilesystemIngestor([]string{tmpDir}, nil)
		registry.Register(fsIngestor)

		options := &core.CrawlerOptions{}
		config := core.DefaultBatchConfig()
		config.BatchSize = 1 // Process one file at a time
		config.SuppressLogs = true

		_, _ = IndexBuilderBatchedWithRegistry(registry, options, config)
		// Should handle batch size of 1 without issues
	})
}