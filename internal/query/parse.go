package query

import (
	"mneme/internal/core"
	"mneme/internal/index"
	"mneme/internal/logger"
)

// ParseQuery tokenizes a query string into stemmed tokens for BM25/VSM scoring.
// This is a convenience wrapper that maintains backward compatibility.
func ParseQuery(queryString string) []string {
	logger.Debug("Tokenizing query")
	// Use the same tokenization pipeline as indexing for BM25 consistency
	tokens := index.TokenizeQuery(queryString)

	return tokens
}

// FindQueryToken finds documents matching the given tokens using BM25+VSM ranking.
func FindQueryToken(segment *core.Segment, tokens []string) []string {
	logger.Info("Finding query token in segments using BM25+VSM ranking")

	// Use the ranking system that combines BM25 and VSM scores
	// Limit results to top MaxResults documents for performance
	return GetTopDocumentPaths(segment, tokens, MaxResults)
}
