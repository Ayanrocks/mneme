package ingest

import (
	"log"
	"mneme/internal/core"
	"mneme/internal/storage"
)

// FilesystemIngestor implements the Ingestor interface for local filesystem sources.
// It wraps the existing storage.Crawler and storage.ReadFileContents functions.
type FilesystemIngestor struct {
	// paths are the root paths to crawl
	paths []string

	// config holds the source-specific configuration
	config *core.FilesystemSourceConfig
}

// NewFilesystemIngestor creates a new filesystem ingestor with the given paths and config.
func NewFilesystemIngestor(paths []string, config *core.FilesystemSourceConfig) *FilesystemIngestor {
	return &FilesystemIngestor{
		paths:  paths,
		config: config,
	}
}

// Name returns "filesystem" as the source identifier.
func (f *FilesystemIngestor) Name() string {
	return "filesystem"
}

// Crawl uses the existing storage.Crawler to discover files.
func (f *FilesystemIngestor) Crawl(options *core.CrawlerOptions) ([]string, error) {
	if options == nil {
		defaultOpts := core.DefaultCrawlerOptions()
		options = &defaultOpts
	}

	allFiles := make([]string, 0)

	for _, path := range f.paths {
		crawlPaths, err := storage.Crawler(path, *options)
		if err != nil {
			log.Printf("Error crawling path %s: %v", path, err)
			continue
		}
		allFiles = append(allFiles, crawlPaths...)
	}

	return allFiles, nil
}

// Read uses storage.ReadFileContents to read a file and wraps it in a Document.
func (f *FilesystemIngestor) Read(id string) (*Document, error) {
	contents, err := storage.ReadFileContents(id)
	if err != nil {
		return nil, err
	}

	return &Document{
		ID:       id,
		Path:     id,
		Contents: contents,
		Source:   f.Name(),
	}, nil
}

// IsEnabled returns true if the filesystem source is enabled in config.
// Filesystem is enabled by default if no config is provided.
func (f *FilesystemIngestor) IsEnabled() bool {
	if f.config == nil {
		return true // Default enabled
	}
	return f.config.Enabled
}
