package query

import (
	"mneme/internal/core"
	"sort"
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

// RankDocuments sorts documents by their combined BM25+VSM score
// and returns the top N results
func RankDocuments(segment *core.Segment, tokens []string, limit int) []RankedDocument {
	if segment == nil || len(tokens) == 0 {
		return []RankedDocument{}
	}

	// Calculate BM25 scores
	bm25Scores := CalculateBM25Scores(segment, tokens)

	// Calculate VSM scores
	vsmScores := CalculateVSMScores(segment, tokens)

	// Combine scores
	combinedScores := CombineScores(bm25Scores, vsmScores, BM25Weight, VSMWeight)

	// Convert to ranked documents
	ranked := make([]RankedDocument, 0, len(combinedScores))

	// Build document ID to path map
	docPaths := make(map[uint]string)
	for _, doc := range segment.Docs {
		docPaths[doc.ID] = doc.Path
	}

	for docID, score := range combinedScores {
		if score > 0 {
			ranked = append(ranked, RankedDocument{
				DocID: docID,
				Path:  docPaths[docID],
				Score: score,
			})
		}
	}

	// Sort by score (descending)
	sort.Slice(ranked, func(i, j int) bool {
		return ranked[i].Score > ranked[j].Score
	})

	// Apply limit
	if limit > 0 && len(ranked) > limit {
		ranked = ranked[:limit]
	}

	return ranked
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
