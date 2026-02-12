package query

import (
	"math"
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

	// Per-token fuzzy expansion: only expand tokens that have no exact match
	// in the inverted index. Track which original tokens map to which expansions
	// so we can calculate fuzzy-aware term coverage later.
	searchTokens := tokens
	// tokenExpansions maps each original token to the set of index terms that represent it.
	// For exact-match tokens, it maps to itself. For fuzzy-expanded tokens, it maps to
	// the fuzzy matches (e.g., "fnid" → ["find"]).
	tokenExpansions := make(map[string][]string, len(tokens))
	for _, token := range tokens {
		tokenExpansions[token] = []string{token} // default: exact match
	}

	vocabulary := GetVocabulary(segment)
	if len(vocabulary) > 0 {
		var missingTokens []string
		for _, token := range tokens {
			if _, exists := segment.InvertedIndex[token]; !exists {
				missingTokens = append(missingTokens, token)
			}
		}
		if len(missingTokens) > 0 {
			fuzzyMatches := ExpandTokensWithFuzzy(missingTokens, vocabulary)
			if len(fuzzyMatches) > 0 {
				searchTokens = MergeTokensWithFuzzy(tokens, fuzzyMatches)
				// Update token expansions for fuzzy-matched tokens
				for _, match := range fuzzyMatches {
					tokenExpansions[match.Original] = append(tokenExpansions[match.Original], match.Matched)
				}
			}
		}
	}

	// Calculate BM25 scores
	bm25Scores := CalculateBM25Scores(segment, searchTokens)

	// Calculate VSM scores
	vsmScores := CalculateVSMScores(segment, searchTokens)

	// Combine scores
	combinedScores := CombineScores(bm25Scores, vsmScores, bm25Weight, vsmWeight)

	// Build document ID to path map
	docPaths := make(map[uint]string)
	for _, doc := range segment.Docs {
		docPaths[doc.ID] = doc.Path
	}

	// Calculate fuzzy-aware term coverage: for each original query token, check if the doc
	// contains ANY of its expansions (exact or fuzzy). This way, a doc matching "find"
	// (the fuzzy expansion of "fnid") counts as covering the "fnid" slot.
	termCoverage := CalculateFuzzyTermCoverage(segment, tokens, tokenExpansions)

	// Build candidate documents with positive scores, applying coverage boost
	candidates := make([]core.RankedDocument, 0, len(combinedScores))
	for docID, score := range combinedScores {
		if score > 0 {
			coverage := termCoverage[docID]
			// Apply coverage boost: score * coverage^1.5
			// This smoothly rewards higher coverage without being too harsh on partial matches
			boostedScore := score * coverage * math.Sqrt(coverage)
			candidates = append(candidates, core.RankedDocument{
				DocID: docID,
				Path:  docPaths[docID],
				Score: boostedScore,
			})
		}
	}

	// Use heap-based TopK to get the highest scoring documents
	return utils.TopK(candidates, limit)
}

// CalculateFuzzyTermCoverage computes what fraction of original query tokens each document covers,
// taking fuzzy expansions into account. For each original token, any of its expansions (exact or
// fuzzy-matched) can satisfy the coverage. Returns a map from docID to coverage ratio (0.0 to 1.0).
func CalculateFuzzyTermCoverage(segment *core.Segment, originalTokens []string, tokenExpansions map[string][]string) map[uint]float64 {
	coverage := make(map[uint]float64)
	if segment == nil || len(originalTokens) == 0 {
		return coverage
	}

	tokenCount := float64(len(originalTokens))

	// For each document, count how many original tokens it covers
	// A document covers an original token if it contains ANY of that token's expansions
	docCoveredTokens := make(map[uint]int)

	for _, origToken := range originalTokens {
		expansions := tokenExpansions[origToken]

		// Collect all doc IDs that contain any expansion of this token
		coveredDocs := make(map[uint]bool)
		for _, expansion := range expansions {
			postings, exists := segment.InvertedIndex[expansion]
			if !exists {
				continue
			}
			for _, posting := range postings {
				coveredDocs[posting.DocID] = true
			}
		}

		// Each doc that covers this token gets +1
		for docID := range coveredDocs {
			docCoveredTokens[docID]++
		}
	}

	for docID, hits := range docCoveredTokens {
		coverage[docID] = float64(hits) / tokenCount
	}

	return coverage
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
