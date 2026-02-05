package query

import (
	"mneme/internal/core"
	"testing"
)

func TestRankDocuments(t *testing.T) {
	t.Run("nil segment returns empty slice", func(t *testing.T) {
		result := RankDocuments(nil, []string{"test"}, 10, nil)
		if len(result) != 0 {
			t.Errorf("Expected empty slice, got %v", result)
		}
	})

	t.Run("empty tokens returns empty slice", func(t *testing.T) {
		segment := createTestSegment()
		result := RankDocuments(segment, []string{}, 10, nil)
		if len(result) != 0 {
			t.Errorf("Expected empty slice, got %v", result)
		}
	})

	t.Run("zero or negative limit uses default", func(t *testing.T) {
		segment := createTestSegment()
		result := RankDocuments(segment, []string{"user"}, 0, nil)
		// Should still return results (using MaxResults as default)
		if len(result) == 0 {
			t.Error("Expected results with default limit")
		}
	})

	t.Run("valid ranking with matching tokens", func(t *testing.T) {
		segment := createTestSegment()
		result := RankDocuments(segment, []string{"user"}, 10, nil)

		// Should have results for docs containing "user"
		if len(result) == 0 {
			t.Error("Expected ranked documents")
		}

		// Results should be sorted by score (descending)
		for i := 1; i < len(result); i++ {
			if result[i].Score > result[i-1].Score {
				t.Errorf("Results not sorted: score[%d]=%v > score[%d]=%v",
					i, result[i].Score, i-1, result[i-1].Score)
			}
		}

		// All results should have positive scores
		for _, doc := range result {
			if doc.Score <= 0 {
				t.Errorf("Doc %d has non-positive score: %v", doc.DocID, doc.Score)
			}
		}
	})

	t.Run("limit restricts result count", func(t *testing.T) {
		segment := createTestSegment()
		result := RankDocuments(segment, []string{"data"}, 2, nil)

		// "data" is in all 3 docs, but limit is 2
		if len(result) > 2 {
			t.Errorf("Expected at most 2 results, got %d", len(result))
		}
	})

	t.Run("results have correct paths", func(t *testing.T) {
		segment := createTestSegment()
		result := RankDocuments(segment, []string{"user"}, 10, nil)

		for _, doc := range result {
			if doc.Path == "" {
				t.Errorf("Doc %d has empty path", doc.DocID)
			}
		}
	})
}

func TestRankDocuments_WithConfig(t *testing.T) {
	t.Run("nil config uses defaults", func(t *testing.T) {
		segment := createTestSegment()
		result := RankDocuments(segment, []string{"user"}, 10, nil)
		if len(result) == 0 {
			t.Error("Expected results with nil config")
		}
	})

	t.Run("custom weights from config", func(t *testing.T) {
		segment := createTestSegment()

		// Use BM25-heavy weighting
		bm25Config := &core.RankingConfig{BM25Weight: 1.0, VSMWeight: 0.0}
		bm25Result := RankDocuments(segment, []string{"user"}, 10, bm25Config)

		// Use VSM-heavy weighting
		vsmConfig := &core.RankingConfig{BM25Weight: 0.0, VSMWeight: 1.0}
		vsmResult := RankDocuments(segment, []string{"user"}, 10, vsmConfig)

		// Both should return results
		if len(bm25Result) == 0 || len(vsmResult) == 0 {
			t.Error("Expected results with custom configs")
		}
	})

	t.Run("invalid config weights use defaults", func(t *testing.T) {
		segment := createTestSegment()

		// Zero weights should trigger default fallback
		config := &core.RankingConfig{BM25Weight: 0, VSMWeight: 0}
		result := RankDocuments(segment, []string{"user"}, 10, config)

		// Should still return results using default weights
		if len(result) == 0 {
			t.Error("Expected results with zero-weight config (fallback to defaults)")
		}
	})

	t.Run("negative weights use defaults", func(t *testing.T) {
		segment := createTestSegment()

		config := &core.RankingConfig{BM25Weight: -1, VSMWeight: -1}
		result := RankDocuments(segment, []string{"user"}, 10, config)

		if len(result) == 0 {
			t.Error("Expected results with negative-weight config (fallback to defaults)")
		}
	})
}

func TestGetTopDocumentPaths(t *testing.T) {
	t.Run("nil segment returns empty slice", func(t *testing.T) {
		result := GetTopDocumentPaths(nil, []string{"test"}, 10)
		if len(result) != 0 {
			t.Errorf("Expected empty slice, got %v", result)
		}
	})

	t.Run("empty tokens returns empty slice", func(t *testing.T) {
		segment := createTestSegment()
		result := GetTopDocumentPaths(segment, []string{}, 10)
		if len(result) != 0 {
			t.Errorf("Expected empty slice, got %v", result)
		}
	})

	t.Run("returns paths only", func(t *testing.T) {
		segment := createTestSegment()
		result := GetTopDocumentPaths(segment, []string{"user"}, 10)

		// Should return paths, not full RankedDocument objects
		for _, path := range result {
			if path == "" {
				t.Error("Expected non-empty paths")
			}
			// Paths should look like file paths
			if path[0] != '/' {
				t.Errorf("Expected absolute path, got %s", path)
			}
		}
	})

	t.Run("limit restricts result count", func(t *testing.T) {
		segment := createTestSegment()
		result := GetTopDocumentPaths(segment, []string{"data"}, 1)

		if len(result) > 1 {
			t.Errorf("Expected at most 1 result, got %d", len(result))
		}
	})
}

func TestRankedDocument_GetScore(t *testing.T) {
	doc := core.RankedDocument{
		DocID: 1,
		Path:  "/test/path",
		Score: 0.85,
	}

	if doc.GetScore() != 0.85 {
		t.Errorf("GetScore() = %v, expected 0.85", doc.GetScore())
	}
}
