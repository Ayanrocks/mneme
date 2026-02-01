package query

import (
	"math"
	"mneme/internal/core"
)

// BM25 parameters - standard values from literature
const (
	// k1 controls term frequency saturation
	// Higher values give more weight to term frequency
	k1 = 1.5

	// b controls document length normalization
	// b=1 means full length normalization, b=0 means no normalization
	b = 0.75
)

// DocumentScore holds a document with its relevance score
type DocumentScore struct {
	DocID uint
	Path  string
	Score float64
}

// calculateIDF calculates the Inverse Document Frequency for a term
// Formula: ln((N - df + 0.5) / (df + 0.5) + 1)
// Where N is total documents, df is document frequency (docs containing term)
func calculateIDF(df int, totalDocs int) float64 {
	if df <= 0 || totalDocs <= 0 {
		return 0
	}

	n := float64(totalDocs)
	docFreq := float64(df)

	// BM25 IDF formula with smoothing to avoid negative values
	return math.Log((n-docFreq+0.5)/(docFreq+0.5) + 1)
}

// calculateTermBM25 calculates BM25 score for a single term in a document
// Formula: IDF × (tf × (k1 + 1)) / (tf + k1 × (1 - b + b × (docLen / avgDocLen)))
func calculateTermBM25(tf float64, idf float64, docLen float64, avgDocLen float64) float64 {
	if avgDocLen <= 0 {
		avgDocLen = 1 // Prevent division by zero
	}

	// Length normalization factor
	lengthNorm := 1 - b + b*(docLen/avgDocLen)

	// BM25 term score
	numerator := tf * (k1 + 1)
	denominator := tf + k1*lengthNorm

	return idf * (numerator / denominator)
}

// CalculateBM25Scores computes BM25 relevance scores for all documents
// against the given query tokens
func CalculateBM25Scores(segment *core.Segment, tokens []string) map[uint]float64 {
	scores := make(map[uint]float64)

	if segment == nil || len(tokens) == 0 {
		return scores
	}

	totalDocs := int(segment.TotalDocs)
	avgDocLen := float64(segment.AvgDocLen)

	// Handle empty index
	if totalDocs == 0 {
		return scores
	}

	// If avgDocLen is 0, calculate it from documents
	if avgDocLen == 0 && len(segment.Docs) > 0 {
		var totalTokens uint
		for _, doc := range segment.Docs {
			totalTokens += doc.TokenCount
		}
		avgDocLen = float64(totalTokens) / float64(len(segment.Docs))
		if avgDocLen == 0 {
			avgDocLen = 1
		}
	}

	// Build document length map for quick lookup
	docLengths := make(map[uint]float64)
	for _, doc := range segment.Docs {
		docLengths[doc.ID] = float64(doc.TokenCount)
	}

	// Process each query token
	for _, token := range tokens {
		postings, exists := segment.InvertedIndex[token]
		if !exists {
			continue
		}

		// Calculate IDF for this term
		df := len(postings)
		idf := calculateIDF(df, totalDocs)

		// Score each document containing this term
		for _, posting := range postings {
			tf := float64(posting.Freq)
			docLen := docLengths[posting.DocID]

			// Accumulate BM25 score for this term
			termScore := calculateTermBM25(tf, idf, docLen, avgDocLen)
			scores[posting.DocID] += termScore
		}
	}

	return scores
}

// GetDocumentFrequency returns how many documents contain a given term
func GetDocumentFrequency(segment *core.Segment, token string) int {
	if segment == nil {
		return 0
	}

	postings, exists := segment.InvertedIndex[token]
	if !exists {
		return 0
	}

	return len(postings)
}

// GetTermFrequencyInDoc returns the frequency of a term in a specific document
func GetTermFrequencyInDoc(segment *core.Segment, token string, docID uint) uint {
	if segment == nil {
		return 0
	}

	postings, exists := segment.InvertedIndex[token]
	if !exists {
		return 0
	}

	for _, posting := range postings {
		if posting.DocID == docID {
			return posting.Freq
		}
	}

	return 0
}
