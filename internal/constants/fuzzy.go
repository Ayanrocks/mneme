package constants

const (
	// FuzzyMaxEditDistance is the maximum Levenshtein distance allowed
	// for a term to be considered a fuzzy match.
	FuzzyMaxEditDistance = 2

	// FuzzyMinTermLength is the minimum length a query term must have
	// before fuzzy matching is applied. Short terms produce too many
	// false positives.
	FuzzyMinTermLength = 4

	// TrigramSimilarityThreshold is the minimum trigram overlap (Dice coefficient)
	// required for a term to be considered a candidate for fuzzy matching.
	TrigramSimilarityThreshold = 0.3

	// FuzzyScorePenalty is applied to fuzzy-matched tokens to prefer exact matches.
	// A value of 0.8 means fuzzy matches contribute 80% of the score weight.
	FuzzyScorePenalty = 0.8
)
