package storage

import (
	"errors"
	"fmt"
	"mneme/internal/core"
	"mneme/internal/logger"
	"mneme/internal/utils"
	"os"
	"path/filepath"
	"strings"
)

// BinaryExtensions is a comprehensive list of file extensions for binary/non-readable files
// that should be skipped during indexing
var BinaryExtensions = map[string]bool{
	// Compiled/binary files
	"exe": true, "dll": true, "so": true, "dylib": true, "o": true, "a": true, "ko": true,
	"pyc": true, "pyo": true, "class": true, "jar": true, "war": true, "ear": true,
	"obj": true, "lib": true, "pdb": true, "ilk": true, "exp": true,

	// Archives
	"zip": true, "tar": true, "gz": true, "bz2": true, "xz": true, "7z": true, "rar": true,
	"iso": true, "dmg": true, "cab": true, "lz": true, "lzma": true, "lzo": true, "z": true,
	"tgz": true, "tbz2": true, "txz": true, "tlz": true,

	// Images
	"jpg": true, "jpeg": true, "png": true, "gif": true, "bmp": true, "ico": true, "svg": true,
	"webp": true, "tiff": true, "tif": true, "raw": true, "psd": true, "ai": true, "eps": true,
	"heic": true, "heif": true, "avif": true, "jfif": true, "exr": true, "dds": true,
	"cr2": true, "nef": true, "orf": true, "sr2": true, "arw": true,

	// Videos
	"mp4": true, "avi": true, "mkv": true, "mov": true, "wmv": true, "flv": true, "webm": true,
	"m4v": true, "mpeg": true, "mpg": true, "3gp": true, "3g2": true, "vob": true, "ogv": true,
	"mxf": true, "mts": true, "m2ts": true, "ts": true, "divx": true, "xvid": true,

	// Audio
	"mp3": true, "wav": true, "flac": true, "aac": true, "ogg": true, "wma": true, "m4a": true,
	"aiff": true, "aif": true, "mid": true, "midi": true, "opus": true, "ape": true, "alac": true,
	"ac3": true, "dts": true, "ra": true, "ram": true,

	// Documents (binary formats)
	"pdf": true, "doc": true, "docx": true, "xls": true, "xlsx": true, "ppt": true, "pptx": true,
	"odt": true, "ods": true, "odp": true, "rtf": true, "wpd": true, "pages": true,
	"numbers": true, "key": true, "epub": true, "mobi": true,

	// Database files
	"db": true, "sqlite": true, "sqlite3": true, "mdb": true, "accdb": true, "dbf": true,
	"frm": true, "myd": true, "myi": true, "ibd": true, "ldf": true, "mdf": true, "ndf": true,

	// Fonts
	"ttf": true, "otf": true, "woff": true, "woff2": true, "eot": true, "fon": true, "fnt": true,

	// Other binary/non-readable files
	"bin": true, "dat": true, "pak": true, "wasm": true, "node": true, "deb": true, "rpm": true,
	"msi": true, "pkg": true, "app": true, "apk": true, "ipa": true, "xap": true,
	"swf": true, "fla": true, "blend": true, "fbx": true, "3ds": true, "max": true,
	"unity": true, "unitypackage": true, "asset": true, "prefab": true,
	"lock": true, "DS_Store": true, "localized": true, "map.js": true,
}

// Crawler crawls the given path and returns a list of file paths
// If the path is a file, it returns that single file path
// If the path is a directory, it recursively crawls all files and nested folders
func Crawler(inputPath string, options core.CrawlerOptions) ([]string, error) {
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
	includeExtMap := buildExtensionMap(options.IncludeExtensions)
	excludeExtMap := buildExtensionMap(options.ExcludeExtensions)
	skipFolderMap := buildFolderMap(options.SkipFolders)

	err = crawlDirectory(expandedPath, &results, includeExtMap, excludeExtMap, skipFolderMap, options)
	if err != nil {
		logger.Errorf("Error during crawl: %+v", err)
		return nil, err
	}

	logger.Debugf("Crawl completed. Found %d files", len(results))
	return results, nil
}

// crawlDirectory recursively crawls a directory and appends file paths to results
func crawlDirectory(dirPath string, results *[]string, includeExtMap map[string]bool, excludeExtMap map[string]bool, skipFolderMap map[string]bool, options core.CrawlerOptions) error {
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
			err := crawlDirectory(entryPath, results, includeExtMap, excludeExtMap, skipFolderMap, options)
			if err != nil {
				// Log the error but continue with other entries
				logger.Warnf("Error crawling subdirectory %s: %+v", entryPath, err)
			}
		} else {
			// It's a file, check if it should be included based on extension
			ext := getFileExtension(entryName)

			// First check if extension is in exclude list
			if excludeExtMap[ext] {
				logger.Debugf("Skipping file due to exclude extension filter: %s", entryPath)
				continue
			}

			// Check if we should skip binary files
			if options.SkipBinaryFiles && BinaryExtensions[ext] {
				logger.Debugf("Skipping binary file: %s", entryPath)
				continue
			}

			// If includeExtMap is non-empty, only include files with matching extensions
			// If includeExtMap is empty, include all files
			if len(includeExtMap) > 0 && !includeExtMap[ext] {
				logger.Debugf("Skipping file due to include extension filter: %s", entryPath)
				continue
			}

			*results = append(*results, entryPath)
		}
	}

	return nil
}

// shouldSkipFile checks if a file should be skipped based on options
func shouldSkipFile(filePath string, options core.CrawlerOptions) bool {
	fileName := filepath.Base(filePath)

	// Check hidden files
	if !options.IncludeHidden && strings.HasPrefix(fileName, ".") {
		return true
	}

	ext := getFileExtension(fileName)

	// Check if extension is in exclude list
	for _, excludeExt := range options.ExcludeExtensions {
		if strings.EqualFold(ext, excludeExt) {
			return true
		}
	}

	// Check if we should skip binary files
	if options.SkipBinaryFiles && BinaryExtensions[ext] {
		return true
	}

	// Check extension - if IncludeExtensions is non-empty, only include matching extensions
	if len(options.IncludeExtensions) > 0 {
		found := false
		for _, includeExt := range options.IncludeExtensions {
			if strings.EqualFold(ext, includeExt) {
				found = true
				break
			}
		}
		if !found {
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
