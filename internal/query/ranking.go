package query

import (
	"math"
	"mneme/internal/constants"
	"mneme/internal/core"
	"path/filepath"
	"sort"
)

// MaxResults is the default limit for search results
const MaxResults = 20

// Default weights for combining BM25 and VSM scores (used when config values are invalid)
const (
	DefaultBM25Weight = 0.7 // Primary relevance signal
	DefaultVSMWeight  = 0.3 // Similarity refinement
)

// RankDocuments uses parallel execution to score documents based on exact and fuzzy matches.
// It merges the results, sums the scores for documents found in both passes, and returns the top K.
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

	// Build document ID to path map upfront for efficient lookup
	docPaths := make(map[uint]string, len(segment.Docs))
	for _, doc := range segment.Docs {
		docPaths[doc.ID] = doc.Path
	}

	// Channels for results
	type searchResult struct {
		docs []core.RankedDocument
	}
	exactCh := make(chan searchResult)
	fuzzyCh := make(chan searchResult)

	// G1: Exact Search
	go func() {
		docs := performExactSearch(segment, tokens, bm25Weight, vsmWeight, docPaths)
		exactCh <- searchResult{docs: docs}
	}()

	// G2: Fuzzy Search
	go func() {
		docs := performFuzzySearch(segment, tokens, bm25Weight, vsmWeight, docPaths)
		fuzzyCh <- searchResult{docs: docs}
	}()

	// Gather results
	exactRes := <-exactCh
	fuzzyRes := <-fuzzyCh

	// Merge results
	// Map docID -> RankedDocument
	mergedDocs := make(map[uint]core.RankedDocument)

	// Helper to merge
	merge := func(docs []core.RankedDocument) {
		for _, doc := range docs {
			if existing, ok := mergedDocs[doc.DocID]; ok {
				existing.Score += doc.Score
				existing.MatchCount += doc.MatchCount
				// Union of matched terms
				existing.MatchedTerms = append(existing.MatchedTerms, doc.MatchedTerms...)
				mergedDocs[doc.DocID] = existing
			} else {
				mergedDocs[doc.DocID] = doc
			}
		}
	}

	merge(exactRes.docs)
	merge(fuzzyRes.docs)

	// Convert to slice for sorting
	finalCandidates := make([]core.RankedDocument, 0, len(mergedDocs))
	for _, doc := range mergedDocs {
		if doc.Score > 0 {
			finalCandidates = append(finalCandidates, doc)
		}
	}

	// Sort with Tie-Breaking Logic
	// 1. Score (Descending)
	// 2. MatchCount (Descending)
	// 3. Filename (Ascending)
	sort.Slice(finalCandidates, func(i, j int) bool {
		d1 := finalCandidates[i]
		d2 := finalCandidates[j]

		// 1. Score
		if math.Abs(d1.Score-d2.Score) > 1e-6 {
			return d1.Score > d2.Score
		}

		// 2. MatchCount
		if d1.MatchCount != d2.MatchCount {
			return d1.MatchCount > d2.MatchCount
		}

		// 3. Filename (Alphabetical Ascending, so "a" before "b" -> i < j)
		// We need to extract filename from path
		f1 := filepath.Base(d1.Path)
		f2 := filepath.Base(d2.Path)
		return f1 < f2 // Ascending
	})

	// Return Top K
	if len(finalCandidates) > limit {
		return finalCandidates[:limit]
	}
	return finalCandidates
}

func performExactSearch(segment *core.Segment, tokens []string, bm25Weight, vsmWeight float64, docPaths map[uint]string) []core.RankedDocument {
	// 1. BM25
	bm25Scores := CalculateBM25Scores(segment, tokens)

	// 2. VSM
	// We use standard VSM here.
	vsmScores := CalculateVSMScores(segment, tokens)

	// 3. Combine
	combined := CombineScores(bm25Scores, vsmScores, bm25Weight, vsmWeight)

	return convertScoresToRankedDocs(segment, combined, tokens, docPaths)
}

func performFuzzySearch(segment *core.Segment, tokens []string, bm25Weight, vsmWeight float64, docPaths map[uint]string) []core.RankedDocument {
	vocabulary := GetVocabulary(segment)
	if len(vocabulary) == 0 {
		return nil
	}

	// 1. Identify and expand tokens
	fuzzyMatches := ExpandTokensWithFuzzy(tokens, vocabulary)
	if len(fuzzyMatches) == 0 {
		return nil
	}

	// Extract unique fuzzy terms (excluding the original tokens if they appeared in exact search)
	fuzzyTermsMap := make(map[string]bool)
	for _, match := range fuzzyMatches {
		// If match.Matched == match.Original, it's an exact match (if distance 0).
		// We skip it because G1 handled it.
		if match.Distance == 0 && match.Matched == match.Original {
			continue
		}
		// If the matched term exists in the original query tokens, skip it (it's handled by G1)
		isOriginal := false
		for _, t := range tokens {
			if t == match.Matched {
				isOriginal = true
				break
			}
		}
		if !isOriginal {
			fuzzyTermsMap[match.Matched] = true
		}
	}

	if len(fuzzyTermsMap) == 0 {
		return nil
	}

	fuzzyTerms := make([]string, 0, len(fuzzyTermsMap))
	for term := range fuzzyTermsMap {
		fuzzyTerms = append(fuzzyTerms, term)
	}

	// 2. Score these fuzzy terms
	bm25Scores := CalculateBM25Scores(segment, fuzzyTerms)
	vsmScores := CalculateVSMScores(segment, fuzzyTerms)

	// Key Step: Apply Penalty to Fuzzy Scores
	for docID := range bm25Scores {
		bm25Scores[docID] *= constants.FuzzyScorePenalty
	}
	for docID := range vsmScores {
		vsmScores[docID] *= constants.FuzzyScorePenalty
	}

	combined := CombineScores(bm25Scores, vsmScores, bm25Weight, vsmWeight)

	return convertScoresToRankedDocs(segment, combined, fuzzyTerms, docPaths)
}

func convertScoresToRankedDocs(segment *core.Segment, scores map[uint]float64, terms []string, docPaths map[uint]string) []core.RankedDocument {
	docs := make([]core.RankedDocument, 0, len(scores))

	for docID, score := range scores {
		if score <= 0 {
			continue
		}

		// Calculate MatchCount
		matchCount := 0
		for _, term := range terms {
			tf := GetTermFrequencyInDoc(segment, term, docID)
			matchCount += int(tf)
		}

		docPath := docPaths[docID] // Use the map for O(1) correct lookup

		docs = append(docs, core.RankedDocument{
			DocID:        docID,
			Path:         docPath,
			Score:        score,
			MatchCount:   matchCount,
			MatchedTerms: terms,
		})
	}
	return docs
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
