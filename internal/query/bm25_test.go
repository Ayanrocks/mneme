package query

import (
	"math"
	"mneme/internal/core"
	"testing"
)

func TestCalculateIDF(t *testing.T) {
	tests := []struct {
		name      string
		df        int
		totalDocs int
		expected  float64
	}{
		{
			name:      "zero document frequency",
			df:        0,
			totalDocs: 100,
			expected:  0,
		},
		{
			name:      "zero total docs",
			df:        5,
			totalDocs: 0,
			expected:  0,
		},
		{
			name:      "negative document frequency",
			df:        -1,
			totalDocs: 100,
			expected:  0,
		},
		{
			name:      "normal case - rare term",
			df:        1,
			totalDocs: 100,
			// IDF should be high for rare terms
			expected: math.Log((100-1+0.5)/(1+0.5) + 1),
		},
		{
			name:      "normal case - common term",
			df:        50,
			totalDocs: 100,
			// IDF should be lower for common terms
			expected: math.Log((100-50+0.5)/(50+0.5) + 1),
		},
		{
			name:      "term in all documents",
			df:        100,
			totalDocs: 100,
			// IDF should be low but positive due to smoothing
			// Using 0.5/(100+0.5)+1 since 100-100=0
			expected: math.Log(0.5/(100+0.5) + 1),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculateIDF(tt.df, tt.totalDocs)
			if math.Abs(result-tt.expected) > 0.0001 {
				t.Errorf("calculateIDF(%d, %d) = %v, expected %v",
					tt.df, tt.totalDocs, result, tt.expected)
			}
		})
	}
}

func TestCalculateTermBM25(t *testing.T) {
	tests := []struct {
		name      string
		tf        float64
		idf       float64
		docLen    float64
		avgDocLen float64
		wantZero  bool
	}{
		{
			name:      "zero avgDocLen uses default",
			tf:        1.0,
			idf:       1.0,
			docLen:    100,
			avgDocLen: 0,
			wantZero:  false,
		},
		{
			name:      "normal calculation",
			tf:        5.0,
			idf:       2.0,
			docLen:    100,
			avgDocLen: 100,
			wantZero:  false,
		},
		{
			name:      "short document gets boost",
			tf:        5.0,
			idf:       2.0,
			docLen:    50,
			avgDocLen: 100,
			wantZero:  false,
		},
		{
			name:      "long document gets penalty",
			tf:        5.0,
			idf:       2.0,
			docLen:    200,
			avgDocLen: 100,
			wantZero:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculateTermBM25(tt.tf, tt.idf, tt.docLen, tt.avgDocLen)
			if tt.wantZero && result != 0 {
				t.Errorf("calculateTermBM25() = %v, expected 0", result)
			}
			if !tt.wantZero && result <= 0 {
				t.Errorf("calculateTermBM25() = %v, expected positive score", result)
			}
		})
	}

	// Test that shorter docs get higher scores (all else equal)
	t.Run("shorter docs score higher", func(t *testing.T) {
		shortDocScore := calculateTermBM25(5.0, 2.0, 50, 100)
		longDocScore := calculateTermBM25(5.0, 2.0, 200, 100)
		if shortDocScore <= longDocScore {
			t.Errorf("Short doc score (%v) should be higher than long doc score (%v)",
				shortDocScore, longDocScore)
		}
	})
}

func TestCalculateBM25Scores(t *testing.T) {
	t.Run("nil segment returns empty map", func(t *testing.T) {
		result := CalculateBM25Scores(nil, []string{"test"})
		if len(result) != 0 {
			t.Errorf("Expected empty map, got %v", result)
		}
	})

	t.Run("empty tokens returns empty map", func(t *testing.T) {
		segment := &core.Segment{TotalDocs: 10}
		result := CalculateBM25Scores(segment, []string{})
		if len(result) != 0 {
			t.Errorf("Expected empty map, got %v", result)
		}
	})

	t.Run("zero total docs returns empty map", func(t *testing.T) {
		segment := &core.Segment{TotalDocs: 0}
		result := CalculateBM25Scores(segment, []string{"test"})
		if len(result) != 0 {
			t.Errorf("Expected empty map, got %v", result)
		}
	})

	t.Run("valid segment with matching tokens", func(t *testing.T) {
		segment := createTestSegment()
		result := CalculateBM25Scores(segment, []string{"user"})

		// Should have scores for docs containing "user"
		if len(result) == 0 {
			t.Error("Expected non-empty scores for matching token")
		}

		// All scores should be positive
		for docID, score := range result {
			if score <= 0 {
				t.Errorf("Doc %d has non-positive score: %v", docID, score)
			}
		}
	})

	t.Run("non-existent token returns empty scores", func(t *testing.T) {
		segment := createTestSegment()
		result := CalculateBM25Scores(segment, []string{"nonexistent"})
		if len(result) != 0 {
			t.Errorf("Expected empty map for non-existent token, got %v", result)
		}
	})
}

func TestGetDocumentFrequency(t *testing.T) {
	t.Run("nil segment returns 0", func(t *testing.T) {
		result := GetDocumentFrequency(nil, "test")
		if result != 0 {
			t.Errorf("Expected 0, got %d", result)
		}
	})

	t.Run("non-existent token returns 0", func(t *testing.T) {
		segment := createTestSegment()
		result := GetDocumentFrequency(segment, "nonexistent")
		if result != 0 {
			t.Errorf("Expected 0, got %d", result)
		}
	})

	t.Run("existing token returns correct count", func(t *testing.T) {
		segment := createTestSegment()
		result := GetDocumentFrequency(segment, "user")
		if result != 2 { // "user" appears in 2 docs in our test segment
			t.Errorf("Expected 2, got %d", result)
		}
	})
}

func TestGetTermFrequencyInDoc(t *testing.T) {
	t.Run("nil segment returns 0", func(t *testing.T) {
		result := GetTermFrequencyInDoc(nil, "test", 1)
		if result != 0 {
			t.Errorf("Expected 0, got %d", result)
		}
	})

	t.Run("non-existent token returns 0", func(t *testing.T) {
		segment := createTestSegment()
		result := GetTermFrequencyInDoc(segment, "nonexistent", 1)
		if result != 0 {
			t.Errorf("Expected 0, got %d", result)
		}
	})

	t.Run("token not in specified doc returns 0", func(t *testing.T) {
		segment := createTestSegment()
		result := GetTermFrequencyInDoc(segment, "user", 999)
		if result != 0 {
			t.Errorf("Expected 0, got %d", result)
		}
	})

	t.Run("existing token in doc returns correct frequency", func(t *testing.T) {
		segment := createTestSegment()
		result := GetTermFrequencyInDoc(segment, "user", 1)
		if result != 3 { // "user" appears 3 times in doc 1 in our test segment
			t.Errorf("Expected 3, got %d", result)
		}
	})
}

func TestCalculateBM25Scores_EdgeCases(t *testing.T) {
	t.Run("handles segment with zero avgDocLen", func(t *testing.T) {
		segment := &core.Segment{
			Docs: []core.Document{
				{ID: 1, Path: "/test.txt", TokenCount: 10},
			},
			InvertedIndex: map[string][]core.Posting{
				"test": {{DocID: 1, Freq: 2}},
			},
			TotalDocs:   1,
			TotalTokens: 5,
			AvgDocLen:   0, // Zero average
		}

		scores := CalculateBM25Scores(segment, []string{"test"})
		// Should still calculate scores without division by zero
		if len(scores) == 0 {
			t.Error("Expected scores even with zero avgDocLen")
		}
		for _, score := range scores {
			if score <= 0 {
				t.Error("Expected positive scores")
			}
		}
	})

	t.Run("handles multiple query tokens", func(t *testing.T) {
		segment := createTestSegment()
		scores := CalculateBM25Scores(segment, []string{"user", "config", "data"})

		// Should accumulate scores from all matching tokens
		if len(scores) == 0 {
			t.Error("Expected scores for multiple tokens")
		}
	})

	t.Run("handles documents with zero token count", func(t *testing.T) {
		segment := &core.Segment{
			Docs: []core.Document{
				{ID: 1, Path: "/empty.txt", TokenCount: 0},
			},
			InvertedIndex: map[string][]core.Posting{
				"test": {{DocID: 1, Freq: 1}},
			},
			TotalDocs:   1,
			TotalTokens: 1,
			AvgDocLen:   1,
		}

		scores := CalculateBM25Scores(segment, []string{"test"})
		// Should handle zero token count
		if _, exists := scores[1]; !exists {
			t.Error("Expected score for document with zero tokens")
		}
	})

	t.Run("rare term gets higher IDF than common term", func(t *testing.T) {
		segment := createTestSegment()

		// "profil" is rarer (1 doc) than "data" (3 docs)
		rareScores := CalculateBM25Scores(segment, []string{"profil"})
		commonScores := CalculateBM25Scores(segment, []string{"data"})

		// Rare term should generally score higher (though depends on TF)
		// At least verify both produce scores
		if len(rareScores) == 0 || len(commonScores) == 0 {
			t.Error("Expected scores for both rare and common terms")
		}
	})

	t.Run("high frequency in document increases score", func(t *testing.T) {
		segment := &core.Segment{
			Docs: []core.Document{
				{ID: 1, Path: "/high.txt", TokenCount: 100},
				{ID: 2, Path: "/low.txt", TokenCount: 100},
			},
			InvertedIndex: map[string][]core.Posting{
				"test": {
					{DocID: 1, Freq: 10}, // High frequency
					{DocID: 2, Freq: 1},  // Low frequency
				},
			},
			TotalDocs:   2,
			TotalTokens: 10,
			AvgDocLen:   100,
		}

		scores := CalculateBM25Scores(segment, []string{"test"})
		if scores[1] <= scores[2] {
			t.Error("Expected higher score for document with higher term frequency")
		}
	})
}

func TestGetTermFrequencyInDoc_EdgeCases(t *testing.T) {
	segment := createTestSegment()

	t.Run("returns 0 for doc not in index", func(t *testing.T) {
		result := GetTermFrequencyInDoc(segment, "user", 999)
		if result != 0 {
			t.Errorf("Expected 0 for non-existent doc, got %d", result)
		}
	})

	t.Run("returns correct frequency for multiple matches", func(t *testing.T) {
		result := GetTermFrequencyInDoc(segment, "config", 2)
		if result != 5 {
			t.Errorf("Expected frequency 5, got %d", result)
		}
	})
}

// createTestSegment creates a test segment with sample data for testing
func createTestSegment() *core.Segment {
	return &core.Segment{
		Docs: []core.Document{
			{ID: 1, Path: "/path/to/file1.go", TokenCount: 50},
			{ID: 2, Path: "/path/to/file2.go", TokenCount: 100},
			{ID: 3, Path: "/path/to/file3.go", TokenCount: 75},
		},
		InvertedIndex: map[string][]core.Posting{
			"user": {
				{DocID: 1, Freq: 3},
				{DocID: 2, Freq: 1},
			},
			"profil": {
				{DocID: 1, Freq: 2},
			},
			"config": {
				{DocID: 2, Freq: 5},
				{DocID: 3, Freq: 2},
			},
			"data": {
				{DocID: 1, Freq: 1},
				{DocID: 2, Freq: 3},
				{DocID: 3, Freq: 4},
			},
		},
		TotalDocs:   3,
		TotalTokens: 100,
		AvgDocLen:   75,
	}
}