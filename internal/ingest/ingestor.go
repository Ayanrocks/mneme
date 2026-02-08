// Package ingest provides a pluggable ingestor pattern for abstracting file sources.
// Sources like filesystem, GitHub, GDrive, OneDrive can implement the Ingestor interface
// to provide documents for indexing.
package ingest

import "mneme/internal/core"

// Document represents a document from any source that can be indexed.
type Document struct {
	// ID is a unique identifier for this document (file path, URL, etc.)
	ID string

	// Path is the display path shown in search results
	Path string

	// Contents is the lines of text content
	Contents []string

	// Source identifies which ingestor provided this document
	Source string
}

// Ingestor defines the interface that all document sources must implement.
// Each source (filesystem, GitHub, GDrive, etc.) provides its own implementation.
type Ingestor interface {
	// Name returns the source name (e.g., "filesystem", "github", "gdrive")
	Name() string

	// Crawl returns all document IDs/paths from this source based on the options.
	// Returns an error if crawling fails.
	Crawl(options *core.CrawlerOptions) ([]string, error)

	// Read retrieves the contents of a single document by its ID.
	// Returns an error if the document cannot be read.
	Read(id string) (*Document, error)

	// IsEnabled checks if this ingestor is enabled in the configuration.
	IsEnabled() bool
}
