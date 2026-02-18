package query

import (
	"strings"

	"mneme/internal/constants"
	"mneme/internal/core"
	"mneme/internal/logger"
	"mneme/internal/utils"

	"github.com/fatih/camelcase"
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
		runeLen := len([]rune(token))
		if runeLen < constants.FuzzyMinTermLength {
			continue
		}

		// Find candidates via trigram similarity
		candidates := utils.FindCandidates(token, trigramIndex, constants.TrigramSimilarityThreshold)

		// Determine max edit distance based on token length (rune count)
		// Dynamic thresholding:
		// - Length < 4: No fuzzy (handled by MinTermLength check above)
		// - Length 4-5: Max distance 1 (prevent "uesr" -> "ue" false positives)
		// - Length >= 6: Max distance 2 (standard fuzzy behavior)
		maxDist := 1
		if runeLen >= 6 {
			maxDist = 2
		}

		// Verify candidates with Damerau-Levenshtein distance
		// (counts transpositions like "uesr"→"user" as a single edit)
		for _, candidate := range candidates {
			if utils.IsWithinDamerauDistance(token, candidate, maxDist) {
				sim := utils.TrigramSimilarity(token, candidate)
				dist := utils.DamerauLevenshteinDistance(token, candidate)
				matches = append(matches, FuzzyMatch{
					Original:   token,
					Matched:    candidate,
					Distance:   dist,
					Similarity: sim,
				})
				logger.Debugf("Fuzzy match: %q → %q (distance=%d, similarity=%.2f)",
					token, candidate, dist, sim)
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

// AutoCorrectQuery attempts to correct typos in raw query terms by finding
// the closest match in the segment's vocabulary.
// Returns the corrected terms and a map of original->corrected for any changes.
func AutoCorrectQuery(segment *core.Segment, terms []string) ([]string, map[string]string) {
	if segment == nil || len(segment.InvertedIndex) == 0 || len(terms) == 0 {
		return terms, nil
	}

	vocabulary := GetVocabulary(segment)
	if len(vocabulary) == 0 {
		return terms, nil
	}

	// Build trigram index for fast fuzzy searching
	trigramIndex := utils.BuildTrigramIndex(vocabulary)

	correctedTerms := make([]string, 0, len(terms))
	corrections := make(map[string]string)

	for _, term := range terms {
		// Try exact match first
		if _, exists := segment.InvertedIndex[term]; exists {
			correctedTerms = append(correctedTerms, term)
			continue
		}

		// Try lowercase
		termLower := strings.ToLower(term)
		if _, exists := segment.InvertedIndex[termLower]; exists {
			correctedTerms = append(correctedTerms, term) // Keep original casing
			if term != termLower {
				corrections[term] = termLower
			}
			continue
		}

		// Not found. Search for correction using lowercase term.
		candidates := utils.FindCandidates(termLower, trigramIndex, constants.TrigramSimilarityThreshold)

		// Fallback: If no trigram candidates for short terms, scan entire vocabulary.
		// This handles cases like "fnid" vs "find" where trigram similarity is too low.
		if len(candidates) == 0 && len(termLower) <= 4 {
			for _, vocabTerm := range vocabulary {
				if utils.IsWithinDamerauDistance(termLower, vocabTerm, constants.FuzzyMaxEditDistance) {
					candidates = append(candidates, vocabTerm)
				}
			}
		}

		bestMatch := ""
		minDist := 100
		maxSim := -1.0

		for _, candidate := range candidates {
			dist := utils.DamerauLevenshteinDistance(termLower, candidate)

			// We only care if dist is within allowed range
			if dist > constants.FuzzyMaxEditDistance {
				continue
			}

			// If strictly better distance, take it
			if dist < minDist {
				minDist = dist
				bestMatch = candidate
				maxSim = utils.TrigramSimilarity(termLower, candidate)
			} else if dist == minDist {
				// Tie-breaker: use trigram similarity
				// Higher similarity is better
				sim := utils.TrigramSimilarity(termLower, candidate)
				if sim > maxSim {
					maxSim = sim
					bestMatch = candidate
				}
			}
		}

		if bestMatch != "" {
			correctedTerms = append(correctedTerms, bestMatch)
			corrections[term] = bestMatch
			logger.Infof("Auto-corrected typo: %q -> %q", term, bestMatch)
		} else {
			// No single-term match found. Try splitting compound terms (e.g. "")
			// Only try this if the term looks like CamelCase (has mix of upper/lower or just not all lower)
			// But simple check: just try camelcase split
			parts := camelcase.Split(term)

			// If split produced multiple parts, verify/correct them individually
			if len(parts) > 1 {
				var correctedParts []string
				var originalParts []string
				allCorrected := true
				wasCorrected := false

				for _, part := range parts {
					partLower := strings.ToLower(part)
					// Skip empty or very short non-alphanumeric parts if they might be noise,
					// but usually camelcase split keeps meaningful parts.

					originalParts = append(originalParts, part)

					// Check exact match for part
					if _, exists := segment.InvertedIndex[partLower]; exists {
						correctedParts = append(correctedParts, partLower)
						allCorrected = false // exact match is not a fuzzy correction
						continue
					}

					// Try fuzzy for part
					// Lower threshold for short parts to ensure we get candidates
					thresh := constants.TrigramSimilarityThreshold
					if len(partLower) <= 4 {
						thresh = 0.01
					}
					partCandidates := utils.FindCandidates(partLower, trigramIndex, thresh)

					// Fallback: If no candidates for short word, scan entire vocabulary
					// This handles cases like "fnid" vs "find" where trigram similarity is 0
					if len(partCandidates) == 0 && len(partLower) <= 4 {
						for _, term := range vocabulary {
							if utils.IsWithinDamerauDistance(partLower, term, constants.FuzzyMaxEditDistance) {
								partCandidates = append(partCandidates, term)
							}
						}
					}

					partBest := ""
					partMinDist := 100
					partMaxScore := -100.0 // Combined score: Sim - (LenDiff * 0.1)

					for _, cand := range partCandidates {
						d := utils.DamerauLevenshteinDistance(partLower, cand)
						// Be stricter with short parts
						threshold := constants.FuzzyMaxEditDistance
						if len(partLower) < 4 {
							threshold = 1
						}

						if d > threshold {
							continue
						}

						// Scoring: Prefer high similarity AND similar length
						// Score = Sim - (LenDiff * 0.1)
						sim := utils.TrigramSimilarity(partLower, cand)
						diff := len(partLower) - len(cand)
						if diff < 0 {
							diff = -diff
						}
						lenDiff := diff
						score := sim - (float64(lenDiff) * 0.1)

						if d < partMinDist {
							partMinDist = d
							partBest = cand
							partMaxScore = score
						} else if d == partMinDist {
							if score > partMaxScore {
								partMaxScore = score
								partBest = cand
							}
						}
					}

					if partBest != "" {
						correctedParts = append(correctedParts, partBest)
						wasCorrected = true
						// allCorrected remains true — this part was fuzzy-corrected
					} else {
						// Fallback to original lowercased — this part was NOT corrected
						correctedParts = append(correctedParts, partLower)
						allCorrected = false
					}
				}

				if wasCorrected && allCorrected {
					// Check if recombining the parts forms a valid term (e.g. "find" + "query" + "token" -> "findquerytoken")
					// This handles cases where a compound identifier was split and corrected, but the user likely
					// meant the specific identifier.
					recombined := strings.Join(correctedParts, "")
					if _, exists := segment.InvertedIndex[recombined]; exists {
						correctedTerms = append(correctedTerms, recombined)
						corrections[term] = recombined
						logger.Infof("Auto-corrected compound typo to combined term: %q -> %q", term, recombined)
						continue
					}

					// If exact match of recombined parts fails (e.g. due to stemming "query"->"queri"),
					// try to find a fuzzy match for the recombined string in the vocabulary.
					// We use a small distance (2) because we expect the parts to be mostly correct.
					var fuzzyMatch string
					minDist := 3 // larger than allowed max

					// Quick check against vocabulary
					// This is O(N) but only happens for corrected compound terms which are rare.
					for _, vocabTerm := range vocabulary {
						dist := utils.DamerauLevenshteinDistance(recombined, vocabTerm)
						if dist <= constants.FuzzyMaxEditDistance && dist < minDist {
							minDist = dist
							fuzzyMatch = vocabTerm
						}
					}

					if fuzzyMatch != "" {
						correctedTerms = append(correctedTerms, fuzzyMatch)
						corrections[term] = fuzzyMatch
						logger.Infof("Auto-corrected compound typo to fuzzy combined term: %q -> %q (dist=%d)", term, fuzzyMatch, minDist)
						continue
					}

					// We successfully processed the parts.
					// Add them to correctedTerms
					correctedTerms = append(correctedTerms, correctedParts...)

					// Record the correction mapping
					correctionStr := strings.Join(correctedParts, " ")
					corrections[term] = correctionStr
					logger.Infof("Auto-corrected compound typo: %q -> %q", term, correctionStr)
					continue
				}
			}

			// Fallback: keep original term
			correctedTerms = append(correctedTerms, term)
		}
	}

	return correctedTerms, corrections
}
