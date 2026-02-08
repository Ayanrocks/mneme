package ingest

import (
	"mneme/internal/core"
	"os"
	"path/filepath"
	"testing"
)

func TestFilesystemIngestor_Name(t *testing.T) {
	ingestor := NewFilesystemIngestor([]string{}, nil)
	if ingestor.Name() != "filesystem" {
		t.Errorf("Expected name 'filesystem', got '%s'", ingestor.Name())
	}
}

func TestFilesystemIngestor_IsEnabled(t *testing.T) {
	t.Run("enabled by default when config is nil", func(t *testing.T) {
		ingestor := NewFilesystemIngestor([]string{}, nil)
		if !ingestor.IsEnabled() {
			t.Error("Expected ingestor to be enabled by default")
		}
	})

	t.Run("enabled when config.Enabled is true", func(t *testing.T) {
		config := &core.FilesystemSourceConfig{Enabled: true}
		ingestor := NewFilesystemIngestor([]string{}, config)
		if !ingestor.IsEnabled() {
			t.Error("Expected ingestor to be enabled")
		}
	})

	t.Run("disabled when config.Enabled is false", func(t *testing.T) {
		config := &core.FilesystemSourceConfig{Enabled: false}
		ingestor := NewFilesystemIngestor([]string{}, config)
		if ingestor.IsEnabled() {
			t.Error("Expected ingestor to be disabled")
		}
	})
}

func TestFilesystemIngestor_Crawl(t *testing.T) {
	// Create temp directory with test files
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.go")
	err := os.WriteFile(testFile, []byte("package main"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	ingestor := NewFilesystemIngestor([]string{tmpDir}, nil)
	options := &core.CrawlerOptions{
		IncludeExtensions: []string{"go"},
	}

	files, err := ingestor.Crawl(options)
	if err != nil {
		t.Fatalf("Crawl error: %v", err)
	}

	if len(files) != 1 {
		t.Errorf("Expected 1 file, got %d", len(files))
	}
}

func TestFilesystemIngestor_Read(t *testing.T) {
	// Create temp file with content
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	content := "line1\nline2\nline3"
	err := os.WriteFile(testFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	ingestor := NewFilesystemIngestor([]string{tmpDir}, nil)
	doc, err := ingestor.Read(testFile)
	if err != nil {
		t.Fatalf("Read error: %v", err)
	}

	if doc.ID != testFile {
		t.Errorf("Expected ID '%s', got '%s'", testFile, doc.ID)
	}

	if doc.Source != "filesystem" {
		t.Errorf("Expected source 'filesystem', got '%s'", doc.Source)
	}

	if len(doc.Contents) != 3 {
		t.Errorf("Expected 3 lines, got %d", len(doc.Contents))
	}
}

func TestRegistry_Register(t *testing.T) {
	registry := NewRegistry()

	ingestor := NewFilesystemIngestor([]string{}, nil)
	registry.Register(ingestor)

	enabled := registry.GetEnabledIngestors()
	if len(enabled) != 1 {
		t.Errorf("Expected 1 enabled ingestor, got %d", len(enabled))
	}
}

func TestRegistry_GetEnabledIngestors(t *testing.T) {
	registry := NewRegistry()

	// Add enabled ingestor
	enabledConfig := &core.FilesystemSourceConfig{Enabled: true}
	enabledIngestor := NewFilesystemIngestor([]string{}, enabledConfig)
	registry.Register(enabledIngestor)

	// Add disabled ingestor (simulated by creating with enabled=false)
	disabledConfig := &core.FilesystemSourceConfig{Enabled: false}
	disabledIngestor := NewFilesystemIngestor([]string{}, disabledConfig)
	registry.Register(disabledIngestor)

	enabled := registry.GetEnabledIngestors()
	if len(enabled) != 1 {
		t.Errorf("Expected 1 enabled ingestor, got %d", len(enabled))
	}
}

func TestRegistry_CrawlAll(t *testing.T) {
	// Create temp directories with test files
	tmpDir1 := t.TempDir()
	tmpDir2 := t.TempDir()

	err := os.WriteFile(filepath.Join(tmpDir1, "file1.go"), []byte("package a"), 0644)
	if err != nil {
		t.Fatal(err)
	}
	err = os.WriteFile(filepath.Join(tmpDir2, "file2.go"), []byte("package b"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	registry := NewRegistry()
	ingestor := NewFilesystemIngestor([]string{tmpDir1, tmpDir2}, nil)
	registry.Register(ingestor)

	options := &core.CrawlerOptions{}
	files, err := registry.CrawlAll(options)
	if err != nil {
		t.Fatalf("CrawlAll error: %v", err)
	}

	if len(files) != 2 {
		t.Errorf("Expected 2 files from CrawlAll, got %d", len(files))
	}
}

func TestRegistry_ReadDocument(t *testing.T) {
	// Create temp file
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	err := os.WriteFile(testFile, []byte("hello\nworld"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	registry := NewRegistry()
	ingestor := NewFilesystemIngestor([]string{tmpDir}, nil)
	registry.Register(ingestor)

	doc, err := registry.ReadDocument(testFile)
	if err != nil {
		t.Fatalf("ReadDocument error: %v", err)
	}

	if doc == nil {
		t.Fatal("Expected document, got nil")
	}

	if len(doc.Contents) != 2 {
		t.Errorf("Expected 2 lines, got %d", len(doc.Contents))
	}
}
