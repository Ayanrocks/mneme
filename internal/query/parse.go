package query

import (
	"mneme/internal/core"
	"mneme/internal/index"
	"mneme/internal/logger"
)

func ParseQuery(queryString string) []string {
	logger.Info("Query string: " + queryString)

	// Use the same tokenization pipeline as indexing for BM25 consistency
	tokens := index.TokenizeQuery(queryString)

	return tokens
}

func FindQueryToken(segment *core.Segment, tokens []string) []string {
	logger.Info("Finding query token in segments using BM25+VSM ranking")

	// Use the ranking system that combines BM25 and VSM scores
	// Limit results to top 20 documents for performance
	return GetTopDocumentPaths(segment, tokens, MaxResults)
}
