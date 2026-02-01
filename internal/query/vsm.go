package query

import (
	"math"
	"mneme/internal/core"
)

// TFIDFVector represents a document or query as a TF-IDF weighted vector
type TFIDFVector struct {
	Weights map[string]float64
	Norm    float64 // L2 norm for cosine similarity
}

// BuildQueryTFIDFVector creates a TF-IDF vector for the query terms
func BuildQueryTFIDFVector(segment *core.Segment, tokens []string) *TFIDFVector {
	vector := &TFIDFVector{
		Weights: make(map[string]float64),
	}

	if segment == nil || len(tokens) == 0 {
		return vector
	}

	totalDocs := int(segment.TotalDocs)
	if totalDocs == 0 {
		return vector
	}

	// Calculate term frequencies in query
	queryTF := make(map[string]int)
	for _, token := range tokens {
		queryTF[token]++
	}

	// Calculate TF-IDF weights for query terms
	var sumSquares float64
	for token, tf := range queryTF {
		// Get document frequency
		df := GetDocumentFrequency(segment, token)
		if df == 0 {
			continue // Skip terms not in index
		}

		// Calculate IDF using the same formula as BM25 for consistency
		idf := calculateIDF(df, totalDocs)

		// TF-IDF weight (using log-normalized TF for query)
		weight := (1 + math.Log(float64(tf))) * idf
		vector.Weights[token] = weight
		sumSquares += weight * weight
	}

	// Calculate L2 norm
	if sumSquares > 0 {
		vector.Norm = math.Sqrt(sumSquares)
	}

	return vector
}

// BuildDocumentTFIDFVector creates a TF-IDF vector for a document
// Only considers the query terms for efficiency
func BuildDocumentTFIDFVector(segment *core.Segment, docID uint, queryTokens []string) *TFIDFVector {
	vector := &TFIDFVector{
		Weights: make(map[string]float64),
	}

	if segment == nil || len(queryTokens) == 0 {
		return vector
	}

	totalDocs := int(segment.TotalDocs)
	if totalDocs == 0 {
		return vector
	}

	var sumSquares float64

	// Only calculate weights for query terms (sparse vector representation)
	for _, token := range queryTokens {
		tf := GetTermFrequencyInDoc(segment, token, docID)
		if tf == 0 {
			continue
		}

		df := GetDocumentFrequency(segment, token)
		if df == 0 {
			continue
		}

		idf := calculateIDF(df, totalDocs)

		// TF-IDF weight using log-normalized TF
		weight := (1 + math.Log(float64(tf))) * idf
		vector.Weights[token] = weight
		sumSquares += weight * weight
	}

	// Calculate L2 norm
	if sumSquares > 0 {
		vector.Norm = math.Sqrt(sumSquares)
	}

	return vector
}

// CalculateCosineSimilarity computes the cosine similarity between two TF-IDF vectors
// Returns a value between 0 (no similarity) and 1 (identical)
func CalculateCosineSimilarity(vec1, vec2 *TFIDFVector) float64 {
	if vec1 == nil || vec2 == nil || vec1.Norm == 0 || vec2.Norm == 0 {
		return 0
	}

	// Calculate dot product (only need to iterate over shared keys)
	var dotProduct float64
	for term, weight1 := range vec1.Weights {
		if weight2, exists := vec2.Weights[term]; exists {
			dotProduct += weight1 * weight2
		}
	}

	// Cosine similarity = dot product / (norm1 * norm2)
	return dotProduct / (vec1.Norm * vec2.Norm)
}

// CalculateVSMScores computes VSM cosine similarity scores for all documents
// that have at least one query term
func CalculateVSMScores(segment *core.Segment, tokens []string) map[uint]float64 {
	scores := make(map[uint]float64)

	if segment == nil || len(tokens) == 0 {
		return scores
	}

	// Build query vector
	queryVector := BuildQueryTFIDFVector(segment, tokens)
	if queryVector.Norm == 0 {
		return scores
	}

	// Find all documents that contain at least one query term
	candidateDocs := make(map[uint]bool)
	for _, token := range tokens {
		postings, exists := segment.InvertedIndex[token]
		if !exists {
			continue
		}
		for _, posting := range postings {
			candidateDocs[posting.DocID] = true
		}
	}

	// Calculate VSM score for each candidate document
	for docID := range candidateDocs {
		docVector := BuildDocumentTFIDFVector(segment, docID, tokens)
		similarity := CalculateCosineSimilarity(queryVector, docVector)
		scores[docID] = similarity
	}

	return scores
}

// CombineScores merges BM25 and VSM scores using weighted average
// The combined score balances relevance (BM25) with similarity (VSM)
func CombineScores(bm25Scores, vsmScores map[uint]float64, bm25Weight, vsmWeight float64) map[uint]float64 {
	combined := make(map[uint]float64)

	// Normalize BM25 scores to [0, 1] range for fair combination
	maxBM25 := 0.0
	for _, score := range bm25Scores {
		if score > maxBM25 {
			maxBM25 = score
		}
	}

	// Combine scores for all documents
	allDocs := make(map[uint]bool)
	for docID := range bm25Scores {
		allDocs[docID] = true
	}
	for docID := range vsmScores {
		allDocs[docID] = true
	}

	for docID := range allDocs {
		bm25 := bm25Scores[docID]
		vsm := vsmScores[docID]

		// Normalize BM25 score
		var normalizedBM25 float64
		if maxBM25 > 0 {
			normalizedBM25 = bm25 / maxBM25
		}

		// Combined score
		combined[docID] = bm25Weight*normalizedBM25 + vsmWeight*vsm
	}

	return combined
}
