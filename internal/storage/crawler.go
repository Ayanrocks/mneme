package storage

import (
	"errors"
	"fmt"
	"mneme/internal/logger"
	"mneme/internal/utils"
	"os"
	"path/filepath"
	"strings"
)

// CrawlerOptions contains configuration for the file crawler
type CrawlerOptions struct {
	// SkipExtensions is a list of file extensions to skip (without the leading dot, e.g., "txt", "log")
	SkipExtensions []string

	// SkipFolders is a list of folder names to skip entirely
	SkipFolders []string

	// MaxFilesPerFolder is the maximum number of files to process per folder
	// If a folder contains more files than this limit, it will be skipped
	// Set to 0 or negative to disable this limit
	MaxFilesPerFolder int

	// IncludeHidden determines whether to include hidden files/folders (those starting with .)
	IncludeHidden bool
}

// DefaultCrawlerOptions returns a CrawlerOptions with sensible defaults
func DefaultCrawlerOptions() CrawlerOptions {
	return CrawlerOptions{
		SkipExtensions:    []string{},
		SkipFolders:       []string{".git", "node_modules", ".svn", ".hg", "__pycache__", ".idea", ".vscode"},
		MaxFilesPerFolder: 0, // No limit by default
		IncludeHidden:     false,
	}
}

// Crawler crawls the given path and returns a list of file paths
// If the path is a file, it returns that single file path
// If the path is a directory, it recursively crawls all files and nested folders
func Crawler(inputPath string, options CrawlerOptions) ([]string, error) {
	expandedPath, err := utils.ExpandFilePath(inputPath)
	if err != nil {
		logger.Errorf("Error expanding path: %+v", err)
		return nil, fmt.Errorf("failed to expand path: %w", err)
	}

	// Get file info to determine if it's a file or directory
	info, err := os.Stat(expandedPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			logger.Errorf("Path does not exist: %s", expandedPath)
			return nil, fmt.Errorf("path does not exist: %s", expandedPath)
		}
		logger.Errorf("Error stating path %s: %+v", expandedPath, err)
		return nil, fmt.Errorf("failed to stat path: %w", err)
	}

	// If it's a file, return it directly (after checking if it should be skipped)
	if !info.IsDir() {
		logger.Debugf("Path is a file, returning single path: %s", expandedPath)

		// Check if this file should be skipped based on extension
		if shouldSkipFile(expandedPath, options) {
			logger.Debugf("File skipped due to extension filter: %s", expandedPath)
			return []string{}, nil
		}

		return []string{expandedPath}, nil
	}

	// It's a directory, perform recursive crawl
	logger.Debugf("Starting crawl of directory: %s", expandedPath)

	var results []string
	skipExtMap := buildExtensionMap(options.SkipExtensions)
	skipFolderMap := buildFolderMap(options.SkipFolders)

	err = crawlDirectory(expandedPath, &results, skipExtMap, skipFolderMap, options)
	if err != nil {
		logger.Errorf("Error during crawl: %+v", err)
		return nil, err
	}

	logger.Debugf("Crawl completed. Found %d files", len(results))
	return results, nil
}

// crawlDirectory recursively crawls a directory and appends file paths to results
func crawlDirectory(dirPath string, results *[]string, skipExtMap map[string]bool, skipFolderMap map[string]bool, options CrawlerOptions) error {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		logger.Errorf("Error reading directory %s: %+v", dirPath, err)
		return fmt.Errorf("failed to read directory %s: %w", dirPath, err)
	}

	// Check if folder exceeds file limit
	if options.MaxFilesPerFolder > 0 {
		fileCount := countFilesInEntries(entries)
		if fileCount > options.MaxFilesPerFolder {
			logger.Debugf("Skipping directory %s: contains %d files, limit is %d", dirPath, fileCount, options.MaxFilesPerFolder)
			return nil
		}
	}

	for _, entry := range entries {
		entryName := entry.Name()
		entryPath := filepath.Join(dirPath, entryName)

		// Skip hidden files/folders if configured
		if !options.IncludeHidden && strings.HasPrefix(entryName, ".") {
			logger.Debugf("Skipping hidden entry: %s", entryPath)
			continue
		}

		if entry.IsDir() {
			// Check if this folder should be skipped
			if skipFolderMap[entryName] {
				logger.Debugf("Skipping folder: %s", entryPath)
				continue
			}

			// Recursively crawl subdirectory
			err := crawlDirectory(entryPath, results, skipExtMap, skipFolderMap, options)
			if err != nil {
				// Log the error but continue with other entries
				logger.Warnf("Error crawling subdirectory %s: %+v", entryPath, err)
			}
		} else {
			// It's a file, check if it should be included
			ext := getFileExtension(entryName)
			if skipExtMap[ext] {
				logger.Debugf("Skipping file due to extension: %s", entryPath)
				continue
			}

			*results = append(*results, entryPath)
		}
	}

	return nil
}

// shouldSkipFile checks if a file should be skipped based on options
func shouldSkipFile(filePath string, options CrawlerOptions) bool {
	fileName := filepath.Base(filePath)

	// Check hidden files
	if !options.IncludeHidden && strings.HasPrefix(fileName, ".") {
		return true
	}

	// Check extension
	ext := getFileExtension(fileName)
	for _, skipExt := range options.SkipExtensions {
		if strings.EqualFold(ext, skipExt) {
			return true
		}
	}

	return false
}

// getFileExtension returns the file extension without the leading dot
func getFileExtension(fileName string) string {
	ext := filepath.Ext(fileName)
	if len(ext) > 0 && ext[0] == '.' {
		return strings.ToLower(ext[1:])
	}
	return strings.ToLower(ext)
}

// buildExtensionMap creates a map for O(1) extension lookups
func buildExtensionMap(extensions []string) map[string]bool {
	m := make(map[string]bool, len(extensions))
	for _, ext := range extensions {
		// Normalize extension (remove leading dot if present, lowercase)
		ext = strings.ToLower(strings.TrimPrefix(ext, "."))
		m[ext] = true
	}
	return m
}

// buildFolderMap creates a map for O(1) folder name lookups
func buildFolderMap(folders []string) map[string]bool {
	m := make(map[string]bool, len(folders))
	for _, folder := range folders {
		m[folder] = true
	}
	return m
}

// countFilesInEntries counts the number of files (non-directories) in a list of entries
func countFilesInEntries(entries []os.DirEntry) int {
	count := 0
	for _, entry := range entries {
		if !entry.IsDir() {
			count++
		}
	}
	return count
}
