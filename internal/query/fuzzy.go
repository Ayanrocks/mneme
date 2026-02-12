package query

import (
	"mneme/internal/constants"
	"mneme/internal/core"
	"mneme/internal/logger"
	"mneme/internal/utils"
)

// FuzzyMatch represents a fuzzy match between a query token and an indexed term.
type FuzzyMatch struct {
	Original   string  // The original query token
	Matched    string  // The matched term from the vocabulary
	Distance   int     // Levenshtein edit distance
	Similarity float64 // Trigram similarity score
}

// GetVocabulary extracts all unique terms from a segment's inverted index.
func GetVocabulary(segment *core.Segment) []string {
	if segment == nil || len(segment.InvertedIndex) == 0 {
		return nil
	}

	vocab := make([]string, 0, len(segment.InvertedIndex))
	for term := range segment.InvertedIndex {
		vocab = append(vocab, term)
	}

	return vocab
}

// ExpandTokensWithFuzzy finds fuzzy matches for query tokens that are long enough
// to benefit from fuzzy matching. Uses trigram similarity for fast candidate filtering
// followed by Levenshtein distance verification.
func ExpandTokensWithFuzzy(tokens []string, vocabulary []string) []FuzzyMatch {
	if len(tokens) == 0 || len(vocabulary) == 0 {
		return nil
	}

	// Build trigram index once for the vocabulary
	trigramIndex := utils.BuildTrigramIndex(vocabulary)

	var matches []FuzzyMatch

	for _, token := range tokens {
		// Skip short tokens — too many false positives
		if len([]rune(token)) < constants.FuzzyMinTermLength {
			continue
		}

		// Find candidates via trigram similarity
		candidates := utils.FindCandidates(token, trigramIndex, constants.TrigramSimilarityThreshold)

		// Verify candidates with Levenshtein distance
		for _, candidate := range candidates {
			if utils.IsWithinEditDistance(token, candidate, constants.FuzzyMaxEditDistance) {
				sim := utils.TrigramSimilarity(token, candidate)
				matches = append(matches, FuzzyMatch{
					Original:   token,
					Matched:    candidate,
					Distance:   utils.LevenshteinDistance(token, candidate),
					Similarity: sim,
				})
				logger.Debugf("Fuzzy match: %q → %q (distance=%d, similarity=%.2f)",
					token, candidate, utils.LevenshteinDistance(token, candidate), sim)
			}
		}
	}

	return matches
}

// MergeTokensWithFuzzy returns the union of original tokens and fuzzy-matched terms,
// with duplicates removed. Original tokens always take precedence.
func MergeTokensWithFuzzy(originalTokens []string, fuzzyMatches []FuzzyMatch) []string {
	seen := make(map[string]bool, len(originalTokens)+len(fuzzyMatches))

	// Add original tokens first
	result := make([]string, 0, len(originalTokens)+len(fuzzyMatches))
	for _, token := range originalTokens {
		if !seen[token] {
			result = append(result, token)
			seen[token] = true
		}
	}

	// Add fuzzy matches (deduped)
	for _, match := range fuzzyMatches {
		if !seen[match.Matched] {
			result = append(result, match.Matched)
			seen[match.Matched] = true
		}
	}

	return result
}
