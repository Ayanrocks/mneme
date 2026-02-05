package query

import (
	"mneme/internal/core"
	"mneme/internal/utils"
)

// MaxResults is the default limit for search results
const MaxResults = 20

// Default weights for combining BM25 and VSM scores (used when config values are invalid)
const (
	DefaultBM25Weight = 0.7 // Primary relevance signal
	DefaultVSMWeight  = 0.3 // Similarity refinement
)

// RankDocuments uses a min-heap to efficiently find the top N documents
// by their combined BM25+VSM score. Time complexity: O(n log k) where k is the limit
// If rankingCfg is nil or contains invalid values, default weights are used.
func RankDocuments(segment *core.Segment, tokens []string, limit int, rankingCfg *core.RankingConfig) []core.RankedDocument {
	if segment == nil || len(tokens) == 0 {
		return []core.RankedDocument{}
	}

	if limit <= 0 {
		limit = MaxResults
	}

	// Use config weights with fallback to defaults if invalid
	bm25Weight := DefaultBM25Weight
	vsmWeight := DefaultVSMWeight
	if rankingCfg != nil {
		if rankingCfg.BM25Weight > 0 {
			bm25Weight = rankingCfg.BM25Weight
		}
		if rankingCfg.VSMWeight > 0 {
			vsmWeight = rankingCfg.VSMWeight
		}
	}

	// Calculate BM25 scores
	bm25Scores := CalculateBM25Scores(segment, tokens)

	// Calculate VSM scores
	vsmScores := CalculateVSMScores(segment, tokens)

	// Combine scores
	combinedScores := CombineScores(bm25Scores, vsmScores, bm25Weight, vsmWeight)

	// Build document ID to path map
	docPaths := make(map[uint]string)
	for _, doc := range segment.Docs {
		docPaths[doc.ID] = doc.Path
	}

	// Build candidate documents with positive scores
	candidates := make([]core.RankedDocument, 0, len(combinedScores))
	for docID, score := range combinedScores {
		if score > 0 {
			candidates = append(candidates, core.RankedDocument{
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
	// Use default weights when called without config
	ranked := RankDocuments(segment, tokens, limit, nil)

	paths := make([]string, len(ranked))
	for i, doc := range ranked {
		paths[i] = doc.Path
	}

	return paths
}
