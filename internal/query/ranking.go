package query

import (
	"mneme/internal/core"
	"mneme/internal/utils"
)

// MaxResults is the default limit for search results
const MaxResults = 20

// Weights for combining BM25 and VSM scores
const (
	BM25Weight = 0.7 // Primary relevance signal
	VSMWeight  = 0.3 // Similarity refinement
)

// RankedDocument represents a search result with its score
type RankedDocument struct {
	DocID uint
	Path  string
	Score float64
}

// GetScore implements utils.Scored interface
func (r RankedDocument) GetScore() float64 {
	return r.Score
}

// RankDocuments uses a min-heap to efficiently find the top N documents
// by their combined BM25+VSM score. Time complexity: O(n log k) where k is the limit
func RankDocuments(segment *core.Segment, tokens []string, limit int) []RankedDocument {
	if segment == nil || len(tokens) == 0 {
		return []RankedDocument{}
	}

	if limit <= 0 {
		limit = MaxResults
	}

	// Calculate BM25 scores
	bm25Scores := CalculateBM25Scores(segment, tokens)

	// Calculate VSM scores
	vsmScores := CalculateVSMScores(segment, tokens)

	// Combine scores
	combinedScores := CombineScores(bm25Scores, vsmScores, BM25Weight, VSMWeight)

	// Build document ID to path map
	docPaths := make(map[uint]string)
	for _, doc := range segment.Docs {
		docPaths[doc.ID] = doc.Path
	}

	// Build candidate documents with positive scores
	candidates := make([]RankedDocument, 0, len(combinedScores))
	for docID, score := range combinedScores {
		if score > 0 {
			candidates = append(candidates, RankedDocument{
				DocID: docID,
				Path:  docPaths[docID],
				Score: score,
			})
		}
	}

	// Use heap-based TopK to get the highest scoring documents
	return utils.TopK(candidates, limit)
}

// GetTopDocumentPaths ranks documents and returns only the file paths
// This is a convenience function for the main search interface
func GetTopDocumentPaths(segment *core.Segment, tokens []string, limit int) []string {
	ranked := RankDocuments(segment, tokens, limit)

	paths := make([]string, len(ranked))
	for i, doc := range ranked {
		paths[i] = doc.Path
	}

	return paths
}
