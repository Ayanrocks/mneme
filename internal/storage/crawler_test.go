package storage

import (
	"mneme/internal/core"
	"os"
	"path/filepath"
	"testing"
)

func TestGetFileExtension(t *testing.T) {
	tests := []struct {
		name     string
		fileName string
		expected string
	}{
		{"simple extension", "file.go", "go"},
		{"uppercase extension", "file.GO", "go"},
		{"mixed case", "file.GoLang", "golang"},
		{"multiple dots", "file.test.go", "go"},
		{"no extension", "Makefile", ""},
		{"hidden file with extension", ".gitignore.bak", "bak"},
		{"hidden file no extension", ".gitignore", "gitignore"},
		{"empty string", "", ""},
		{"just dot", ".", ""},
		{"double dot", "..", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getFileExtension(tt.fileName)
			if result != tt.expected {
				t.Errorf("getFileExtension(%q) = %q, expected %q", tt.fileName, result, tt.expected)
			}
		})
	}
}

func TestBuildExtensionMap(t *testing.T) {
	tests := []struct {
		name       string
		extensions []string
		checkIn    []string
		checkOut   []string
	}{
		{
			name:       "empty list",
			extensions: []string{},
			checkIn:    []string{},
			checkOut:   []string{"go", "py"},
		},
		{
			name:       "simple list",
			extensions: []string{"go", "py", "js"},
			checkIn:    []string{"go", "py", "js"},
			checkOut:   []string{"ts", "rs"},
		},
		{
			name:       "normalizes with dot prefix",
			extensions: []string{".go", ".py"},
			checkIn:    []string{"go", "py"},
			checkOut:   []string{".go"}, // Dot is stripped
		},
		{
			name:       "normalizes uppercase",
			extensions: []string{"GO", "PY", "Js"},
			checkIn:    []string{"go", "py", "js"},
			checkOut:   []string{"GO", "PY"},
		},
		{
			name:       "handles duplicates",
			extensions: []string{"go", "go", "GO"},
			checkIn:    []string{"go"},
			checkOut:   []string{"py"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildExtensionMap(tt.extensions)

			for _, ext := range tt.checkIn {
				if !result[ext] {
					t.Errorf("Expected %q to be in map", ext)
				}
			}

			for _, ext := range tt.checkOut {
				if result[ext] {
					t.Errorf("Expected %q to NOT be in map", ext)
				}
			}
		})
	}
}

func TestBuildFolderMap(t *testing.T) {
	tests := []struct {
		name     string
		folders  []string
		checkIn  []string
		checkOut []string
	}{
		{
			name:     "empty list",
			folders:  []string{},
			checkIn:  []string{},
			checkOut: []string{".git", "node_modules"},
		},
		{
			name:     "common skip folders",
			folders:  []string{".git", "node_modules", ".svn"},
			checkIn:  []string{".git", "node_modules", ".svn"},
			checkOut: []string{"src", "cmd"},
		},
		{
			name:     "preserves case",
			folders:  []string{"MyFolder", "UPPERCASE"},
			checkIn:  []string{"MyFolder", "UPPERCASE"},
			checkOut: []string{"myfolder", "uppercase"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildFolderMap(tt.folders)

			for _, folder := range tt.checkIn {
				if !result[folder] {
					t.Errorf("Expected %q to be in map", folder)
				}
			}

			for _, folder := range tt.checkOut {
				if result[folder] {
					t.Errorf("Expected %q to NOT be in map", folder)
				}
			}
		})
	}
}

func TestDefaultCrawlerOptions(t *testing.T) {
	opts := core.DefaultCrawlerOptions()

	t.Run("has default skip folders", func(t *testing.T) {
		expectedFolders := []string{".git", "node_modules", ".svn", ".hg", "__pycache__"}
		for _, folder := range expectedFolders {
			found := false
			for _, skip := range opts.SkipFolders {
				if skip == folder {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Expected %q in default SkipFolders", folder)
			}
		}
	})

	t.Run("empty include extensions by default", func(t *testing.T) {
		if len(opts.IncludeExtensions) != 0 {
			t.Errorf("Expected empty IncludeExtensions, got %v", opts.IncludeExtensions)
		}
	})

	t.Run("empty exclude extensions by default", func(t *testing.T) {
		if len(opts.ExcludeExtensions) != 0 {
			t.Errorf("Expected empty ExcludeExtensions, got %v", opts.ExcludeExtensions)
		}
	})

	t.Run("hidden files excluded by default", func(t *testing.T) {
		if opts.IncludeHidden {
			t.Error("Expected IncludeHidden to be false by default")
		}
	})

	t.Run("no file limit by default", func(t *testing.T) {
		if opts.MaxFilesPerFolder != 0 {
			t.Errorf("Expected MaxFilesPerFolder to be 0 (no limit), got %d", opts.MaxFilesPerFolder)
		}
	})
}

func TestShouldSkipFile(t *testing.T) {
	tests := []struct {
		name     string
		filePath string
		options  core.CrawlerOptions
		expected bool
	}{
		{
			name:     "skip hidden file",
			filePath: "/path/.hidden",
			options:  core.CrawlerOptions{IncludeHidden: false},
			expected: true,
		},
		{
			name:     "include hidden file when enabled",
			filePath: "/path/.hidden",
			options:  core.CrawlerOptions{IncludeHidden: true},
			expected: false,
		},
		{
			name:     "skip excluded extension",
			filePath: "/path/file.log",
			options:  core.CrawlerOptions{ExcludeExtensions: []string{"log", "tmp"}},
			expected: true,
		},
		{
			name:     "skip binary file when enabled",
			filePath: "/path/file.exe",
			options:  core.CrawlerOptions{SkipBinaryFiles: true},
			expected: true,
		},
		{
			name:     "include binary file when disabled",
			filePath: "/path/file.exe",
			options:  core.CrawlerOptions{SkipBinaryFiles: false},
			expected: false,
		},
		{
			name:     "skip if not in include list",
			filePath: "/path/file.py",
			options:  core.CrawlerOptions{IncludeExtensions: []string{"go", "js"}},
			expected: true,
		},
		{
			name:     "include if in include list",
			filePath: "/path/file.go",
			options:  core.CrawlerOptions{IncludeExtensions: []string{"go", "js"}},
			expected: false,
		},
		{
			name:     "include all when include list empty",
			filePath: "/path/file.xyz",
			options:  core.CrawlerOptions{IncludeExtensions: []string{}},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := shouldSkipFile(tt.filePath, tt.options)
			if result != tt.expected {
				t.Errorf("shouldSkipFile(%q, %+v) = %v, expected %v",
					tt.filePath, tt.options, result, tt.expected)
			}
		})
	}
}

func TestBinaryExtensions(t *testing.T) {
	t.Run("contains image formats", func(t *testing.T) {
		imageExts := []string{"jpg", "jpeg", "png", "gif", "bmp", "webp"}
		for _, ext := range imageExts {
			if !BinaryExtensions[ext] {
				t.Errorf("Expected %q in BinaryExtensions", ext)
			}
		}
	})

	t.Run("contains video formats", func(t *testing.T) {
		videoExts := []string{"mp4", "avi", "mkv", "mov", "wmv"}
		for _, ext := range videoExts {
			if !BinaryExtensions[ext] {
				t.Errorf("Expected %q in BinaryExtensions", ext)
			}
		}
	})

	t.Run("contains audio formats", func(t *testing.T) {
		audioExts := []string{"mp3", "wav", "flac", "aac", "ogg"}
		for _, ext := range audioExts {
			if !BinaryExtensions[ext] {
				t.Errorf("Expected %q in BinaryExtensions", ext)
			}
		}
	})

	t.Run("contains compiled formats", func(t *testing.T) {
		compiledExts := []string{"exe", "dll", "so", "dylib", "pyc", "class"}
		for _, ext := range compiledExts {
			if !BinaryExtensions[ext] {
				t.Errorf("Expected %q in BinaryExtensions", ext)
			}
		}
	})

	t.Run("does not contain source code formats", func(t *testing.T) {
		// Note: "ts" is intentionally NOT in this list because .ts is also a video format (MPEG Transport Stream)
		sourceExts := []string{"go", "py", "js", "rs", "c", "cpp", "java"}
		for _, ext := range sourceExts {
			if BinaryExtensions[ext] {
				t.Errorf("Did not expect %q in BinaryExtensions", ext)
			}
		}
	})
}

func TestCrawler_SingleFile(t *testing.T) {
	// Create a temporary file
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.go")
	err := os.WriteFile(testFile, []byte("package main"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	t.Run("returns single file path", func(t *testing.T) {
		opts := core.DefaultCrawlerOptions()
		result, err := Crawler(testFile, opts)
		if err != nil {
			t.Fatalf("Crawler error: %v", err)
		}
		if len(result) != 1 {
			t.Errorf("Expected 1 file, got %d", len(result))
		}
		if result[0] != testFile {
			t.Errorf("Expected %s, got %s", testFile, result[0])
		}
	})

	t.Run("respects extension filter for single file", func(t *testing.T) {
		opts := core.CrawlerOptions{IncludeExtensions: []string{"py"}}
		result, err := Crawler(testFile, opts)
		if err != nil {
			t.Fatalf("Crawler error: %v", err)
		}
		if len(result) != 0 {
			t.Errorf("Expected 0 files (filtered out), got %d", len(result))
		}
	})
}

func TestCrawler_Directory(t *testing.T) {
	// Create a temporary directory structure
	tmpDir := t.TempDir()

	// Create files
	files := []string{
		"main.go",
		"utils.go",
		"readme.md",
		"config.json",
	}
	for _, f := range files {
		err := os.WriteFile(filepath.Join(tmpDir, f), []byte("content"), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
	}

	// Create subdirectory with file
	subDir := filepath.Join(tmpDir, "pkg")
	err := os.Mkdir(subDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create subdirectory: %v", err)
	}
	err = os.WriteFile(filepath.Join(subDir, "helper.go"), []byte("package pkg"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	t.Run("crawls all files", func(t *testing.T) {
		opts := core.DefaultCrawlerOptions()
		result, err := Crawler(tmpDir, opts)
		if err != nil {
			t.Fatalf("Crawler error: %v", err)
		}
		// 4 files in root + 1 in subdirectory = 5 total
		if len(result) != 5 {
			t.Errorf("Expected 5 files, got %d: %v", len(result), result)
		}
	})

	t.Run("filters by extension", func(t *testing.T) {
		opts := core.CrawlerOptions{IncludeExtensions: []string{"go"}}
		result, err := Crawler(tmpDir, opts)
		if err != nil {
			t.Fatalf("Crawler error: %v", err)
		}
		// main.go, utils.go, helper.go = 3 files
		if len(result) != 3 {
			t.Errorf("Expected 3 .go files, got %d: %v", len(result), result)
		}
	})

	t.Run("excludes extensions", func(t *testing.T) {
		opts := core.CrawlerOptions{ExcludeExtensions: []string{"md", "json"}}
		result, err := Crawler(tmpDir, opts)
		if err != nil {
			t.Fatalf("Crawler error: %v", err)
		}
		// Excludes readme.md and config.json = 3 files
		if len(result) != 3 {
			t.Errorf("Expected 3 files (excluding md/json), got %d: %v", len(result), result)
		}
	})
}

func TestCrawler_SkipFolders(t *testing.T) {
	tmpDir := t.TempDir()

	// Create .git folder (should be skipped)
	gitDir := filepath.Join(tmpDir, ".git")
	err := os.Mkdir(gitDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create .git dir: %v", err)
	}
	err = os.WriteFile(filepath.Join(gitDir, "config"), []byte("git config"), 0644)
	if err != nil {
		t.Fatalf("Failed to create .git/config: %v", err)
	}

	// Create node_modules folder (should be skipped)
	nodeDir := filepath.Join(tmpDir, "node_modules")
	err = os.Mkdir(nodeDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create node_modules dir: %v", err)
	}
	err = os.WriteFile(filepath.Join(nodeDir, "package.json"), []byte("{}"), 0644)
	if err != nil {
		t.Fatalf("Failed to create package.json: %v", err)
	}

	// Create regular file
	err = os.WriteFile(filepath.Join(tmpDir, "main.go"), []byte("package main"), 0644)
	if err != nil {
		t.Fatalf("Failed to create main.go: %v", err)
	}

	opts := core.DefaultCrawlerOptions()
	result, err := Crawler(tmpDir, opts)
	if err != nil {
		t.Fatalf("Crawler error: %v", err)
	}

	// Should only find main.go (skips .git and node_modules)
	if len(result) != 1 {
		t.Errorf("Expected 1 file (skipping .git and node_modules), got %d: %v", len(result), result)
	}
}

func TestCrawler_NonExistentPath(t *testing.T) {
	opts := core.DefaultCrawlerOptions()
	_, err := Crawler("/nonexistent/path/that/does/not/exist", opts)
	if err == nil {
		t.Error("Expected error for non-existent path")
	}
}

func TestCountFilesInEntries(t *testing.T) {
	tmpDir := t.TempDir()

	// Create files and directories
	err := os.WriteFile(filepath.Join(tmpDir, "file1.txt"), []byte("1"), 0644)
	if err != nil {
		t.Fatal(err)
	}
	err = os.WriteFile(filepath.Join(tmpDir, "file2.txt"), []byte("2"), 0644)
	if err != nil {
		t.Fatal(err)
	}
	err = os.Mkdir(filepath.Join(tmpDir, "subdir"), 0755)
	if err != nil {
		t.Fatal(err)
	}

	entries, err := os.ReadDir(tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	count := countFilesInEntries(entries)
	if count != 2 {
		t.Errorf("Expected 2 files, got %d", count)
	}
}
