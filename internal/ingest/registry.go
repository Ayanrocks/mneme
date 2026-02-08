package ingest

import (
	"errors"
	"fmt"
	"mneme/internal/core"
	"mneme/internal/logger"
)

var ErrDocumentNotFound = errors.New("document not found")

// Registry manages multiple ingestors and provides unified access to all sources.
type Registry struct {
	ingestors []Ingestor
}

// NewRegistry creates a new empty registry.
func NewRegistry() *Registry {
	return &Registry{
		ingestors: make([]Ingestor, 0),
	}
}

// Register adds an ingestor to the registry.
func (r *Registry) Register(ingestor Ingestor) {
	r.ingestors = append(r.ingestors, ingestor)
}

// GetEnabledIngestors returns all ingestors that are currently enabled.
func (r *Registry) GetEnabledIngestors() []Ingestor {
	enabled := make([]Ingestor, 0)
	for _, ing := range r.ingestors {
		if ing.IsEnabled() {
			enabled = append(enabled, ing)
		}
	}
	return enabled
}

// CrawlAll crawls all enabled sources and returns combined document IDs.
func (r *Registry) CrawlAll(options *core.CrawlerOptions) ([]string, error) {
	allDocs := make([]string, 0)

	for _, ing := range r.GetEnabledIngestors() {
		logger.Debugf("Crawling source: %s", ing.Name())
		docs, err := ing.Crawl(options)
		if err != nil {
			logger.Errorf("Error crawling source %s: %+v", ing.Name(), err)
			continue
		}
		logger.Debugf("Found %d documents from %s", len(docs), ing.Name())
		allDocs = append(allDocs, docs...)
	}

	return allDocs, nil
}

// ReadDocument finds the appropriate ingestor for a document ID and reads it.
// For now, this iterates through enabled ingestors; future optimization could
// use a prefix/pattern matching approach.
func (r *Registry) ReadDocument(id string) (*Document, error) {
	// For filesystem paths, just use the first enabled ingestor that can read it
	for _, ing := range r.GetEnabledIngestors() {
		doc, err := ing.Read(id)
		if err == nil {
			return doc, nil
		}

		// If the error is something other than "not found", return it immediately
		if !errors.Is(err, ErrDocumentNotFound) {
			return nil, err
		}

		// If it is "not found", continue to the next ingestor
	}
	return nil, fmt.Errorf("%w: %s", ErrDocumentNotFound, id)
}

// GetIngestorForDocument returns the ingestor that can handle this document ID.
// Currently returns the first enabled ingestor (filesystem).
func (r *Registry) GetIngestorForDocument(id string) Ingestor {
	enabledIngestors := r.GetEnabledIngestors()
	if len(enabledIngestors) > 0 {
		return enabledIngestors[0]
	}
	return nil
}
