package query

import (
	"fmt"
	"math"
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

	t.Run("fuzzy match with full coverage beats partial exact match", func(t *testing.T) {
		// Scenario: User searches for "fnid query token"
		// Doc 1 (parse.go): "find" (3), "query" (11), "token" (13). Length 45. Full coverage (fuzzy fnid->find).
		// Doc 2 (noise): "find" (0), "query" (50), "token" (50). Length 1000. Partial coverage.

		segment := &core.Segment{
			Docs: []core.Document{
				{ID: 1, Path: "parse.go", TokenCount: 45},
				{ID: 2, Path: "noise.go", TokenCount: 1000},
				// Add dummy docs to stabilize IDF
				{ID: 3, Path: "d1", TokenCount: 100},
				{ID: 4, Path: "d2", TokenCount: 100},
				{ID: 5, Path: "d3", TokenCount: 100},
			},
			InvertedIndex: map[string][]core.Posting{
				"find": {
					{DocID: 1, Freq: 3},
					{DocID: 3, Freq: 1}, // Make find not unique to doc 1
				},
				"query": {
					{DocID: 1, Freq: 11},
					{DocID: 2, Freq: 50}, // High freq in doc2
					{DocID: 3, Freq: 5},
					{DocID: 4, Freq: 5},
				},
				"token": {
					{DocID: 1, Freq: 13},
					{DocID: 2, Freq: 50}, // High freq in doc2
					{DocID: 3, Freq: 5},
					{DocID: 5, Freq: 5},
				},
			},
			TotalDocs:   5,
			TotalTokens: 1345,
			AvgDocLen:   269,
		}

		// Search for "fnid", "query", "token"
		// "fnid" -> fuzzy matches "find"
		tokens := []string{"fnid", "query", "token"}

		results := RankDocuments(segment, tokens, 10, nil)

		if len(results) < 2 {
			t.Fatalf("Expected at least 2 results, got %d", len(results))
		}

		// Doc 1 should be ranked first due to coverage
		if results[0].DocID != 1 {
			t.Errorf("Expected Doc 1 (parse.go) to rank first, but got Doc %d. Scores: Doc1=%v, Doc2=%v",
				results[0].DocID, results[0].Score, results[1].Score)
			// Print for debugging
			for _, r := range results {
				t.Logf("Doc %d (%s): Score %f", r.DocID, r.Path, r.Score)
			}
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

func TestRankDocuments_ParallelTieBreaking(t *testing.T) {
	// Scenario:
	// Doc 1: "apple banana" (Exact + Exact)
	// Doc 2: "apple banan" (Exact + Fuzzy)
	// Doc 3: "apple" (Exact)
	// Doc 4: "banan" (Fuzzy)
	// Doc 5: "apple banana" (Identical to Doc 1 but different filename "z.txt")
	// Doc 6: "apple banana" (Identical to Doc 1 but different filename "a.txt")

	segment := &core.Segment{
		Docs: []core.Document{
			{ID: 1, Path: "doc1.txt", TokenCount: 2},
			{ID: 2, Path: "doc2.txt", TokenCount: 2},
			{ID: 3, Path: "doc3.txt", TokenCount: 1},
			{ID: 4, Path: "doc4.txt", TokenCount: 1},
			{ID: 5, Path: "z.txt", TokenCount: 2},
			{ID: 6, Path: "a.txt", TokenCount: 2},
		},
		InvertedIndex: map[string][]core.Posting{
			"apple": {
				{DocID: 1, Freq: 1},
				{DocID: 2, Freq: 1},
				{DocID: 3, Freq: 1},
				{DocID: 5, Freq: 1},
				{DocID: 6, Freq: 1},
			},
			"banana": {
				{DocID: 1, Freq: 1},
				{DocID: 5, Freq: 1},
				{DocID: 6, Freq: 1},
			},
			"banan": {
				{DocID: 2, Freq: 1},
				{DocID: 4, Freq: 1},
			},
		},
		TotalDocs:   6,
		TotalTokens: 10,
		AvgDocLen:   2,
	}

	// Query: "apple banana"
	// "apple" is exact.
	// "banana" is exact.
	// But we simulate a typo in query? No, the query is "apple banan".
	//
	// Case 1: Query "apple banan"
	// "apple": exact match
	// "banan": exact match (in docs 2,4) AND fuzzy match to "banana" (in docs 1,5,6).
	//
	// Doc 2 has exact "apple" and exact "banan".
	// Doc 1 has exact "apple" and fuzzy "banan" (via "banana").
	//
	// Wait, if "banan" is in the inverted index, it's an Exact Match.
	// The fuzzy logic only expands tokens that correspond to Valid Vocabulary Terms.
	// "banan" is in vocabulary. "banana" is in vocabulary.
	// "banan" matches "banana" fuzzily? Yes, dist=1.
	//
	// If the user types "banan", and "banan" exists in index, it's an exact match.
	// Will it ALSO be expanded to "banana"?
	// Implementation says:
	// `fuzzyMatches := ExpandTokensWithFuzzy(tokens, vocabulary)`
	// `ExpandTokensWithFuzzy` usually checks if token is in vocab?
	// `ExpandTokensWithFuzzy` logic:
	// for _, token := range tokens {
	//    // Find candidates via trigram...
	//    // ...
	// }
	// It doesn't filter out if token itself is in vocab.
	// AND in `performFuzzySearch`:
	// `fuzzyMatches := ExpandTokensWithFuzzy(tokens, vocabulary)`
	// `fuzzyTermsMap` initialization:
	// `for _, match := range fuzzyMatches {`
	// `   if match.Distance == 0 && match.Matched == match.Original { continue }`
	// `   // if matched term exists in tokens...`
	//
	// So if "banan" is in tokens:
	// 1. Exact Search scores "banan".
	// 2. Fuzzy Search expands "banan" -> "banana".
	//    "banana" is NOT in tokens ["apple", "banan"].
	//    So "banana" is added to fuzzy search.
	//
	// Result:
	// Doc 2 ("apple", "banan"): Exact("apple") + Exact("banan").
	// Doc 1 ("apple", "banana"): Exact("apple") + Fuzzy("banana" from "banan").
	//
	// Doc 1 should span both Exact match for "apple" and Fuzzy match for "banana".
	//
	// Tie-breaking check:
	// Doc 5 and Doc 6 are identical to Doc 1 content-wise.
	// Doc 1, 5, 6 should have identical scores.
	// Sorting should be: "a.txt" (Doc 6) < "doc1.txt" (Doc 1) < "z.txt" (Doc 5).

	tokens := []string{"apple", "banan"}
	results := RankDocuments(segment, tokens, 10, nil)

	if len(results) == 0 {
		t.Fatal("Expected results")
	}

	// Top results should include 1, 2, 5, 6.
	// Doc 2 (Exact+Exact) might higher or lower than Doc 1 (Exact+Fuzzy)?
	// "banan" has df=2. "banana" has df=3.
	// Exact matches are preferred (no penalty).
	// So Doc 2 should likely be highest.

	// Check Doc 6, 1, 5 ordering (Tie-breaking by filename)
	// We need to find their indices.
	var idx1, idx5, idx6 int
	found1, found5, found6 := false, false, false

	for i, doc := range results {
		if doc.DocID == 1 {
			idx1 = i
			found1 = true
		} else if doc.DocID == 5 {
			idx5 = i
			found5 = true
		} else if doc.DocID == 6 {
			idx6 = i
			found6 = true
		}
	}

	if !found1 || !found5 || !found6 {
		t.Fatalf("Docs 1, 5, 6 not found in results: %v", results)
	}

	// Check scores are effectively equal
	score1 := results[idx1].Score
	score5 := results[idx5].Score
	score6 := results[idx6].Score

	if math.Abs(score1-score5) > 1e-6 || math.Abs(score1-score6) > 1e-6 {
		t.Errorf("Docs 1, 5, 6 should have equal scores. Got %v, %v, %v", score1, score5, score6)
	}

	// Precedence check: 6 < 1 < 5 based on filename "a.txt" < "doc1.txt" < "z.txt"
	if idx6 > idx1 {
		t.Errorf("Expected Doc 6 (a.txt) before Doc 1 (doc1.txt), got indices %d, %d.\nDetails:\n%v", idx6, idx1, formatResults(results))
	}
	if idx1 > idx5 {
		t.Errorf("Expected Doc 1 (doc1.txt) before Doc 5 (z.txt), got indices %d, %d.\nDetails:\n%v", idx1, idx5, formatResults(results))
	}
}

func formatResults(results []core.RankedDocument) string {
	s := ""
	for i, r := range results {
		s += fmt.Sprintf("[%d] ID:%d Path:%s Score:%.4f MatchCount:%d\n", i, r.DocID, r.Path, r.Score, r.MatchCount)
	}
	return s
}
