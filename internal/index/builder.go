package index

import (
	"fmt"
	"mneme/internal/core"
	"mneme/internal/ingest"
	"mneme/internal/logger"
	"mneme/internal/storage"
	"path/filepath"
	"time"
)

// DEPRECATED: Use IndexBuilderBatched instead
func IndexBuilder(paths []string, crawlerOptions *core.CrawlerOptions) *core.Segment {
	logger.Info("Starting IndexBuilder")

	// Token frequency map: token -> frequency count
	tokenFrequency := make(map[string]uint)
	invertedIndex := make(map[string][]core.Posting)
	docs := make([]core.Document, 0)
	docId := uint(1)

	// Loop through the path to get the contents first
	for _, path := range paths {
		// Get a new instance of the crawler for the specified path
		crawlPaths, err := storage.Crawler(path, *crawlerOptions)
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

// IndexBuilderBatched processes files in batches to reduce memory usage.
// Each batch is written to disk as a numbered chunk file before processing the next batch.
// Returns a manifest tracking all created chunks.
// DEPRECATED: Use IndexBuilderBatchedWithRegistry for pluggable source support.
func IndexBuilderBatched(paths []string, crawlerOptions *core.CrawlerOptions, config *core.BatchConfig) (*core.Manifest, error) {
	if config == nil {
		config = core.DefaultBatchConfig()
	}

	if !config.SuppressLogs {
		logger.Info("Starting IndexBuilderBatched")
	}

	// First, collect all file paths from all sources
	allFiles := make([]string, 0)
	for _, path := range paths {
		crawlPaths, err := storage.Crawler(path, *crawlerOptions)
		if err != nil {
			logger.Errorf("Error crawling path %s: %+v", path, err)
			continue
		}
		allFiles = append(allFiles, crawlPaths...)
	}

	if !config.SuppressLogs {
		logger.Infof("Total files to index: %d (batch size: %d)", len(allFiles), config.BatchSize)
	}

	if len(allFiles) == 0 {
		logger.Warn("No files found to index")
		return nil, nil
	}

	// Initialize manifest
	manifest := core.NewManifest()
	chunkID := 1
	globalDocID := uint(1)

	// Process files in batches
	for batchStart := 0; batchStart < len(allFiles); batchStart += config.BatchSize {
		batchEnd := batchStart + config.BatchSize
		if batchEnd > len(allFiles) {
			batchEnd = len(allFiles)
		}

		batchFiles := allFiles[batchStart:batchEnd]
		if !config.SuppressLogs {
			logger.Infof("Processing batch %d: files %d-%d", chunkID, batchStart+1, batchEnd)
		}

		// Report progress if callback is provided
		if config.ProgressCallback != nil {
			config.ProgressCallback(batchEnd, len(allFiles), fmt.Sprintf("Processing batch %d: files %d-%d", chunkID, batchStart+1, batchEnd))
		}

		// Process this batch
		chunk, docCount, tokenCount := processBatch(batchFiles, &globalDocID)

		// Add chunk info to manifest (marked as in_progress)
		chunkInfo := core.ChunkInfo{
			ID:         chunkID,
			Filename:   formatChunkFilename(chunkID),
			Status:     core.ChunkStatusInProgress,
			DocCount:   docCount,
			TokenCount: tokenCount,
			CreatedAt:  time.Now(),
		}
		manifest.AddChunk(chunkInfo)

		// Save chunk to disk
		err := storage.SaveChunk(chunk, chunkID)
		if err != nil {
			logger.Errorf("Error saving chunk %d: %+v", chunkID, err)
			return manifest, err
		}

		// Mark chunk as complete
		manifest.MarkChunkComplete(chunkID)

		// Save manifest after each chunk (for crash recovery)
		manifest.UpdateTotals()
		err = storage.SaveManifest(manifest)
		if err != nil {
			logger.Errorf("Error saving manifest: %+v", err)
			return manifest, err
		}

		if !config.SuppressLogs {
			logger.Infof("Batch %d completed: %d docs, %d unique tokens", chunkID, docCount, tokenCount)
		}

		// Report batch completion if callback is provided
		if config.ProgressCallback != nil {
			config.ProgressCallback(batchEnd, len(allFiles), fmt.Sprintf("Batch %d completed: %d docs, %d unique tokens", chunkID, docCount, tokenCount))
		}

		chunkID++
	}

	if !config.SuppressLogs {
		logger.Infof("IndexBuilderBatched completed: %d chunks, %d total docs, %d total tokens",
			len(manifest.Chunks), manifest.TotalDocs, manifest.TotalTokens)
	}

	return manifest, nil
}

// IndexBuilderBatchedWithRegistry processes documents from the ingestor registry in batches.
// This is the preferred method for indexing as it supports pluggable sources.
func IndexBuilderBatchedWithRegistry(registry *ingest.Registry, crawlerOptions *core.CrawlerOptions, config *core.BatchConfig) (*core.Manifest, error) {
	if config == nil {
		config = core.DefaultBatchConfig()
	}

	if !config.SuppressLogs {
		logger.Info("Starting IndexBuilderBatchedWithRegistry")
	}

	// Crawl all enabled sources
	allDocIDs, err := registry.CrawlAll(crawlerOptions)
	if err != nil {
		logger.Errorf("Error crawling sources: %+v", err)
		return nil, err
	}

	if !config.SuppressLogs {
		logger.Infof("Total documents to index: %d (batch size: %d)", len(allDocIDs), config.BatchSize)
	}

	if len(allDocIDs) == 0 {
		logger.Warn("No documents found to index")
		return nil, nil
	}

	// Initialize manifest
	manifest := core.NewManifest()
	chunkID := 1
	globalDocID := uint(1)

	// Process documents in batches
	for batchStart := 0; batchStart < len(allDocIDs); batchStart += config.BatchSize {
		batchEnd := batchStart + config.BatchSize
		if batchEnd > len(allDocIDs) {
			batchEnd = len(allDocIDs)
		}

		batchDocIDs := allDocIDs[batchStart:batchEnd]
		if !config.SuppressLogs {
			logger.Infof("Processing batch %d: documents %d-%d", chunkID, batchStart+1, batchEnd)
		}

		// Report progress if callback is provided
		if config.ProgressCallback != nil {
			config.ProgressCallback(batchEnd, len(allDocIDs), fmt.Sprintf("Processing batch %d: documents %d-%d", chunkID, batchStart+1, batchEnd))
		}

		// Process this batch using the registry
		chunk, docCount, tokenCount := processBatchWithRegistry(batchDocIDs, registry, &globalDocID)

		// Add chunk info to manifest (marked as in_progress)
		chunkInfo := core.ChunkInfo{
			ID:         chunkID,
			Filename:   formatChunkFilename(chunkID),
			Status:     core.ChunkStatusInProgress,
			DocCount:   docCount,
			TokenCount: tokenCount,
			CreatedAt:  time.Now(),
		}
		manifest.AddChunk(chunkInfo)

		// Save chunk to disk
		err := storage.SaveChunk(chunk, chunkID)
		if err != nil {
			logger.Errorf("Error saving chunk %d: %+v", chunkID, err)
			return manifest, err
		}

		// Mark chunk as complete
		manifest.MarkChunkComplete(chunkID)

		// Save manifest after each chunk (for crash recovery)
		manifest.UpdateTotals()
		err = storage.SaveManifest(manifest)
		if err != nil {
			logger.Errorf("Error saving manifest: %+v", err)
			return manifest, err
		}

		if !config.SuppressLogs {
			logger.Infof("Batch %d completed: %d docs, %d unique tokens", chunkID, docCount, tokenCount)
		}

		// Report batch completion if callback is provided
		if config.ProgressCallback != nil {
			config.ProgressCallback(batchEnd, len(allDocIDs), fmt.Sprintf("Batch %d completed: %d docs, %d unique tokens", chunkID, docCount, tokenCount))
		}

		chunkID++
	}

	if !config.SuppressLogs {
		logger.Infof("IndexBuilderBatchedWithRegistry completed: %d chunks, %d total docs, %d total tokens",
			len(manifest.Chunks), manifest.TotalDocs, manifest.TotalTokens)
	}

	return manifest, nil
}

// processBatch processes a batch of files and returns a segment chunk
func processBatch(files []string, globalDocID *uint) (*core.Segment, uint, uint) {
	tokenFrequency := make(map[string]uint)
	invertedIndex := make(map[string][]core.Posting)
	docs := make([]core.Document, 0, len(files))
	docCount := uint(0)

	for _, filePath := range files {
		// Reset token frequency for each file
		clear(tokenFrequency)

		fileContents, err := storage.ReadFileContents(filePath)
		if err != nil {
			logger.Errorf("Error reading file %s: %+v", filePath, err)
			continue
		}

		for _, content := range fileContents {
			tokens := Tokenize(content)
			for _, token := range tokens {
				tokenFrequency[token]++
			}
		}

		// Build inverted index for this document
		for token, frequency := range tokenFrequency {
			invertedIndex[token] = append(invertedIndex[token], core.Posting{
				DocID: *globalDocID,
				Freq:  frequency,
			})
		}

		docs = append(docs, core.Document{
			ID:         *globalDocID,
			Path:       filepath.Clean(filePath),
			TokenCount: uint(len(tokenFrequency)),
		})

		*globalDocID++
		docCount++
	}

	avgDocLen := uint(0)
	if len(docs) > 0 {
		totalTokens := uint(0)
		for _, doc := range docs {
			totalTokens += doc.TokenCount
		}
		avgDocLen = totalTokens / uint(len(docs))
	}

	chunk := &core.Segment{
		Docs:          docs,
		InvertedIndex: invertedIndex,
		TotalDocs:     docCount,
		TotalTokens:   uint(len(invertedIndex)),
		AvgDocLen:     avgDocLen,
	}

	return chunk, docCount, uint(len(invertedIndex))
}

// formatChunkFilename returns the filename for a chunk (e.g., "001.idx")
func formatChunkFilename(chunkID int) string {
	return fmt.Sprintf("%03d.idx", chunkID)
}

// processBatchWithRegistry processes a batch of document IDs using the ingestor registry
func processBatchWithRegistry(docIDs []string, registry *ingest.Registry, globalDocID *uint) (*core.Segment, uint, uint) {
	tokenFrequency := make(map[string]uint)
	invertedIndex := make(map[string][]core.Posting)
	docs := make([]core.Document, 0, len(docIDs))
	docCount := uint(0)

	for _, docID := range docIDs {
		// Reset token frequency for each document
		clear(tokenFrequency)

		// Read document via the registry
		doc, err := registry.ReadDocument(docID)
		if err != nil || doc == nil {
			logger.Errorf("Error reading document %s: %+v", docID, err)
			continue
		}

		for _, content := range doc.Contents {
			tokens := Tokenize(content)
			for _, token := range tokens {
				tokenFrequency[token]++
			}
		}

		// Build inverted index for this document
		for token, frequency := range tokenFrequency {
			invertedIndex[token] = append(invertedIndex[token], core.Posting{
				DocID: *globalDocID,
				Freq:  frequency,
			})
		}

		docs = append(docs, core.Document{
			ID:         *globalDocID,
			Path:       filepath.Clean(doc.Path),
			TokenCount: uint(len(tokenFrequency)),
		})

		*globalDocID++
		docCount++
	}

	avgDocLen := uint(0)
	if len(docs) > 0 {
		totalTokens := uint(0)
		for _, doc := range docs {
			totalTokens += doc.TokenCount
		}
		avgDocLen = totalTokens / uint(len(docs))
	}

	chunk := &core.Segment{
		Docs:          docs,
		InvertedIndex: invertedIndex,
		TotalDocs:     docCount,
		TotalTokens:   uint(len(invertedIndex)),
		AvgDocLen:     avgDocLen,
	}

	return chunk, docCount, uint(len(invertedIndex))
}

// Tokenize takes file content as a string and returns a slice of normalized tokens.
// It uses the generic tokenizer which supports camelCase, snake_case, kebab-case
// identifiers, applies Porter stemming for BM25 consistency, and handles binary detection.
func Tokenize(content string) []string {
	return TokenizeContent(content)
}
