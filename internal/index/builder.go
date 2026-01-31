package index

import (
	"mneme/internal/core"
	"mneme/internal/logger"
	"mneme/internal/storage"
	"path/filepath"
)

func IndexBuilder(paths []string) *core.Segment {
	logger.Info("Starting IndexBuilder")

	// Token frequency map: token -> frequency count
	tokenFrequency := make(map[string]uint)
	invertedIndex := make(map[string][]core.Posting)
	docs := make([]core.Document, 0)
	docId := uint(1)

	// Loop through the path to get the contents first
	for _, path := range paths {
		// Get a new instance of the crawler for the specified path
		crawlPaths, err := storage.Crawler(path, storage.DefaultCrawlerOptions())
		if err != nil {
			logger.Errorf("Error crawling path %s: %+v", path, err)
			continue
		}

		logger.Debugf("Crawled %d files", len(crawlPaths))

		for _, filePath := range crawlPaths {
			logger.Debugf("Tokenizing file %s, %d", filePath, docId)
			fileContents, err := storage.ReadFileContents(filePath)
			if err != nil {
				logger.Errorf("Error reading file %s: %+v", filePath, err)
				continue
			}

			for _, content := range fileContents {
				tokens := Tokenize(content)

				// Store token frequencies
				for _, token := range tokens {
					tokenFrequency[token]++
				}
			}

			// After the token map is created, we need to create the inverted index
			for token, frequency := range tokenFrequency {
				invertedIndex[token] = append(invertedIndex[token], core.Posting{
					DocID: docId,
					Freq:  frequency,
				})
			}

			docs = append(docs, core.Document{
				ID:         docId,
				Path:       filepath.Clean(filePath),
				TokenCount: uint(len(tokenFrequency)),
			})

			docId++
		}
	}

	logger.Debugf("Total unique tokens: %d", len(tokenFrequency))
	logger.Info("IndexBuilder completed")

	// return the segment index
	return &core.Segment{
		Docs:          docs,
		InvertedIndex: invertedIndex,
		TotalDocs:     docId,
		TotalTokens:   uint(len(tokenFrequency)),
		AvgDocLen:     uint(len(tokenFrequency) / len(docs)),
	}
}

// Tokenize takes file content as a string and returns a slice of normalized tokens.
// It uses the generic tokenizer which supports camelCase, snake_case, kebab-case
// identifiers, applies Porter stemming for BM25 consistency, and handles binary detection.
func Tokenize(content string) []string {
	return TokenizeContent(content)
}
