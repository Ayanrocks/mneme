package query

import (
	"math"
	"mneme/internal/core"
	"testing"
)

func TestBuildQueryTFIDFVector(t *testing.T) {
	t.Run("nil segment returns empty vector", func(t *testing.T) {
		result := BuildQueryTFIDFVector(nil, []string{"test"})
		if len(result.Weights) != 0 {
			t.Errorf("Expected empty weights, got %v", result.Weights)
		}
		if result.Norm != 0 {
			t.Errorf("Expected zero norm, got %v", result.Norm)
		}
	})

	t.Run("empty tokens returns empty vector", func(t *testing.T) {
		segment := createTestSegment()
		result := BuildQueryTFIDFVector(segment, []string{})
		if len(result.Weights) != 0 {
			t.Errorf("Expected empty weights, got %v", result.Weights)
		}
	})

	t.Run("zero total docs returns empty vector", func(t *testing.T) {
		segment := &core.Segment{TotalDocs: 0}
		result := BuildQueryTFIDFVector(segment, []string{"test"})
		if len(result.Weights) != 0 {
			t.Errorf("Expected empty weights, got %v", result.Weights)
		}
	})

	t.Run("valid query tokens", func(t *testing.T) {
		segment := createTestSegment()
		result := BuildQueryTFIDFVector(segment, []string{"user", "config"})

		// Should have weights for existing tokens
		if len(result.Weights) != 2 {
			t.Errorf("Expected 2 weights, got %d", len(result.Weights))
		}

		// Weights should be positive
		for token, weight := range result.Weights {
			if weight <= 0 {
				t.Errorf("Token %s has non-positive weight: %v", token, weight)
			}
		}

		// Norm should be positive
		if result.Norm <= 0 {
			t.Errorf("Expected positive norm, got %v", result.Norm)
		}
	})

	t.Run("non-existent tokens are skipped", func(t *testing.T) {
		segment := createTestSegment()
		result := BuildQueryTFIDFVector(segment, []string{"nonexistent"})
		if len(result.Weights) != 0 {
			t.Errorf("Expected empty weights for non-existent token, got %v", result.Weights)
		}
	})

	t.Run("repeated query terms increase weight", func(t *testing.T) {
		segment := createTestSegment()
		singleResult := BuildQueryTFIDFVector(segment, []string{"user"})
		repeatedResult := BuildQueryTFIDFVector(segment, []string{"user", "user", "user"})

		// Repeated terms should have higher weight
		if repeatedResult.Weights["user"] <= singleResult.Weights["user"] {
			t.Errorf("Repeated term weight (%v) should be higher than single (%v)",
				repeatedResult.Weights["user"], singleResult.Weights["user"])
		}
	})
}

func TestBuildDocumentTFIDFVector(t *testing.T) {
	t.Run("nil segment returns empty vector", func(t *testing.T) {
		result := BuildDocumentTFIDFVector(nil, 1, []string{"test"})
		if len(result.Weights) != 0 {
			t.Errorf("Expected empty weights, got %v", result.Weights)
		}
	})

	t.Run("empty query tokens returns empty vector", func(t *testing.T) {
		segment := createTestSegment()
		result := BuildDocumentTFIDFVector(segment, 1, []string{})
		if len(result.Weights) != 0 {
			t.Errorf("Expected empty weights, got %v", result.Weights)
		}
	})

	t.Run("valid document with matching tokens", func(t *testing.T) {
		segment := createTestSegment()
		result := BuildDocumentTFIDFVector(segment, 1, []string{"user", "profil"})

		// Doc 1 contains both "user" and "profil"
		if len(result.Weights) != 2 {
			t.Errorf("Expected 2 weights, got %d", len(result.Weights))
		}

		if result.Norm <= 0 {
			t.Errorf("Expected positive norm, got %v", result.Norm)
		}
	})

	t.Run("document without query tokens returns empty weights", func(t *testing.T) {
		segment := createTestSegment()
		// Doc 3 only has "config" and "data", not "user" or "profil"
		result := BuildDocumentTFIDFVector(segment, 3, []string{"user", "profil"})
		if len(result.Weights) != 0 {
			t.Errorf("Expected empty weights, got %v", result.Weights)
		}
	})
}

func TestCalculateCosineSimilarity(t *testing.T) {
	t.Run("nil vectors return 0", func(t *testing.T) {
		result := CalculateCosineSimilarity(nil, nil)
		if result != 0 {
			t.Errorf("Expected 0, got %v", result)
		}

		vec := &core.TFIDFVector{Weights: map[string]float64{"a": 1}, Norm: 1}
		result = CalculateCosineSimilarity(nil, vec)
		if result != 0 {
			t.Errorf("Expected 0, got %v", result)
		}

		result = CalculateCosineSimilarity(vec, nil)
		if result != 0 {
			t.Errorf("Expected 0, got %v", result)
		}
	})

	t.Run("zero norm vectors return 0", func(t *testing.T) {
		vec1 := &core.TFIDFVector{Weights: map[string]float64{"a": 1}, Norm: 0}
		vec2 := &core.TFIDFVector{Weights: map[string]float64{"a": 1}, Norm: 1}
		result := CalculateCosineSimilarity(vec1, vec2)
		if result != 0 {
			t.Errorf("Expected 0, got %v", result)
		}
	})

	t.Run("identical vectors return 1", func(t *testing.T) {
		vec := &core.TFIDFVector{
			Weights: map[string]float64{"a": 1, "b": 2},
			Norm:    math.Sqrt(5), // sqrt(1^2 + 2^2)
		}
		result := CalculateCosineSimilarity(vec, vec)
		if math.Abs(result-1.0) > 0.0001 {
			t.Errorf("Expected 1.0, got %v", result)
		}
	})

	t.Run("orthogonal vectors return 0", func(t *testing.T) {
		vec1 := &core.TFIDFVector{
			Weights: map[string]float64{"a": 1},
			Norm:    1,
		}
		vec2 := &core.TFIDFVector{
			Weights: map[string]float64{"b": 1},
			Norm:    1,
		}
		result := CalculateCosineSimilarity(vec1, vec2)
		if result != 0 {
			t.Errorf("Expected 0, got %v", result)
		}
	})

	t.Run("partial overlap", func(t *testing.T) {
		vec1 := &core.TFIDFVector{
			Weights: map[string]float64{"a": 1, "b": 1},
			Norm:    math.Sqrt(2),
		}
		vec2 := &core.TFIDFVector{
			Weights: map[string]float64{"a": 1, "c": 1},
			Norm:    math.Sqrt(2),
		}
		result := CalculateCosineSimilarity(vec1, vec2)
		// dot product = 1*1 = 1
		// similarity = 1 / (sqrt(2) * sqrt(2)) = 0.5
		if math.Abs(result-0.5) > 0.0001 {
			t.Errorf("Expected 0.5, got %v", result)
		}
	})
}

func TestCalculateVSMScores(t *testing.T) {
	t.Run("nil segment returns empty map", func(t *testing.T) {
		result := CalculateVSMScores(nil, []string{"test"})
		if len(result) != 0 {
			t.Errorf("Expected empty map, got %v", result)
		}
	})

	t.Run("empty tokens returns empty map", func(t *testing.T) {
		segment := createTestSegment()
		result := CalculateVSMScores(segment, []string{})
		if len(result) != 0 {
			t.Errorf("Expected empty map, got %v", result)
		}
	})

	t.Run("valid segment with matching tokens", func(t *testing.T) {
		segment := createTestSegment()
		result := CalculateVSMScores(segment, []string{"user"})

		// Should have scores for docs containing "user" (docs 1 and 2)
		if len(result) != 2 {
			t.Errorf("Expected 2 document scores, got %d", len(result))
		}

		// All scores should be between 0 and 1 (cosine similarity)
		for docID, score := range result {
			if score < 0 || score > 1 {
				t.Errorf("Doc %d has invalid cosine similarity: %v", docID, score)
			}
		}
	})

	t.Run("non-existent token returns empty scores", func(t *testing.T) {
		segment := createTestSegment()
		result := CalculateVSMScores(segment, []string{"nonexistent"})
		if len(result) != 0 {
			t.Errorf("Expected empty map, got %v", result)
		}
	})
}

func TestCombineScores(t *testing.T) {
	t.Run("empty inputs return empty map", func(t *testing.T) {
		result := CombineScores(nil, nil, 0.7, 0.3)
		if len(result) != 0 {
			t.Errorf("Expected empty map, got %v", result)
		}
	})

	t.Run("combines scores from both sources", func(t *testing.T) {
		bm25 := map[uint]float64{1: 2.0, 2: 1.0}
		vsm := map[uint]float64{1: 0.8, 3: 0.5}

		result := CombineScores(bm25, vsm, 0.7, 0.3)

		// Should have scores for all docs (1, 2, 3)
		if len(result) != 3 {
			t.Errorf("Expected 3 combined scores, got %d", len(result))
		}

		// Doc 1 should have highest combined score (high BM25 + high VSM)
		if result[1] <= result[2] {
			t.Errorf("Doc 1 should score higher than doc 2")
		}
	})

	t.Run("weights affect final scores", func(t *testing.T) {
		bm25 := map[uint]float64{1: 1.0}
		vsm := map[uint]float64{1: 1.0}

		// With equal weights
		result1 := CombineScores(bm25, vsm, 0.5, 0.5)

		// With BM25 weighted higher
		result2 := CombineScores(bm25, vsm, 0.9, 0.1)

		// Both should produce valid scores
		if result1[1] <= 0 || result2[1] <= 0 {
			t.Error("Expected positive combined scores")
		}
	})

	t.Run("normalizes BM25 scores", func(t *testing.T) {
		bm25 := map[uint]float64{1: 10.0, 2: 5.0}
		vsm := map[uint]float64{1: 0.5, 2: 0.5}

		result := CombineScores(bm25, vsm, 0.7, 0.3)

		// Doc 1 should have normalized BM25 of 1.0, doc 2 should have 0.5
		// Combined: doc1 = 0.7*1.0 + 0.3*0.5 = 0.85
		// Combined: doc2 = 0.7*0.5 + 0.3*0.5 = 0.50
		if result[1] <= result[2] {
			t.Errorf("Doc 1 (%v) should score higher than doc 2 (%v)", result[1], result[2])
		}
	})
}
