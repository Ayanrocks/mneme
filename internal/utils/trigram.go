package utils

// GenerateTrigrams generates character-level 3-grams from a term with padding.
// The term is padded with '$' characters to capture boundary information.
// Example: "cat" → ["$$c", "$ca", "cat", "at$", "t$$"]
func GenerateTrigrams(term string) []string {
	if len(term) == 0 {
		return nil
	}

	// Pad the term with $ for boundary trigrams
	padded := "$$" + term + "$$"

	var trigrams []string
	runes := []rune(padded)
	for i := 0; i <= len(runes)-3; i++ {
		trigrams = append(trigrams, string(runes[i:i+3]))
	}

	return trigrams
}

// TrigramSimilarity calculates the Dice coefficient between two terms based
// on their trigram sets. Returns a value between 0.0 (no overlap) and 1.0 (identical).
// Dice = 2 * |intersection| / (|A| + |B|)
func TrigramSimilarity(a, b string) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0.0
	}

	if a == b {
		return 1.0
	}

	trigramsA := GenerateTrigrams(a)
	trigramsB := GenerateTrigrams(b)

	if len(trigramsA) == 0 || len(trigramsB) == 0 {
		return 0.0
	}

	// Build set from B's trigrams
	setB := make(map[string]bool, len(trigramsB))
	for _, t := range trigramsB {
		setB[t] = true
	}

	// Count intersection
	intersection := 0
	seen := make(map[string]bool, len(trigramsA))
	for _, t := range trigramsA {
		if setB[t] && !seen[t] {
			intersection++
			seen[t] = true
		}
	}

	// Dice coefficient: 2 * |intersection| / (|A unique| + |B unique|)
	uniqueA := uniqueCount(trigramsA)
	uniqueB := uniqueCount(trigramsB)

	if uniqueA+uniqueB == 0 {
		return 0.0
	}

	return 2.0 * float64(intersection) / float64(uniqueA+uniqueB)
}

// uniqueCount returns the number of unique strings in a slice.
func uniqueCount(strs []string) int {
	seen := make(map[string]bool, len(strs))
	for _, s := range strs {
		seen[s] = true
	}
	return len(seen)
}

// BuildTrigramIndex creates an inverted index mapping trigrams to the terms
// that contain them. This enables fast candidate lookup during fuzzy search.
func BuildTrigramIndex(terms []string) map[string][]string {
	index := make(map[string][]string)

	for _, term := range terms {
		trigrams := GenerateTrigrams(term)
		seen := make(map[string]bool, len(trigrams))
		for _, tri := range trigrams {
			if !seen[tri] {
				index[tri] = append(index[tri], term)
				seen[tri] = true
			}
		}
	}

	return index
}

// FindCandidates returns terms from the trigram index whose trigram similarity
// to the query term meets or exceeds the given threshold.
// The trigram index should be pre-built via BuildTrigramIndex.
func FindCandidates(queryTerm string, trigramIndex map[string][]string, threshold float64) []string {
	queryTrigrams := GenerateTrigrams(queryTerm)
	if len(queryTrigrams) == 0 {
		return nil
	}

	// Collect candidate terms that share at least one trigram
	candidateCounts := make(map[string]int)
	for _, tri := range queryTrigrams {
		if terms, exists := trigramIndex[tri]; exists {
			for _, term := range terms {
				candidateCounts[term]++
			}
		}
	}

	// Filter by actual trigram similarity
	var result []string
	for candidate := range candidateCounts {
		if candidate == queryTerm {
			continue // Skip exact match — handled separately
		}
		sim := TrigramSimilarity(queryTerm, candidate)
		if sim >= threshold {
			result = append(result, candidate)
		}
	}

	return result
}
