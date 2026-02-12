package query

import (
	"mneme/internal/core"
	"testing"
)

func TestGetVocabulary(t *testing.T) {
	t.Run("nil segment returns nil", func(t *testing.T) {
		result := GetVocabulary(nil)
		if result != nil {
			t.Errorf("Expected nil, got %v", result)
		}
	})

	t.Run("empty inverted index returns nil", func(t *testing.T) {
		segment := &core.Segment{
			InvertedIndex: map[string][]core.Posting{},
		}
		result := GetVocabulary(segment)
		if result != nil {
			t.Errorf("Expected nil, got %v", result)
		}
	})

	t.Run("returns all terms from inverted index", func(t *testing.T) {
		segment := createTestSegment()
		vocab := GetVocabulary(segment)

		// Should have all unique terms
		expected := map[string]bool{
			"user":   true,
			"profil": true,
			"config": true,
			"data":   true,
		}

		if len(vocab) != len(expected) {
			t.Errorf("Expected %d terms, got %d: %v", len(expected), len(vocab), vocab)
			return
		}

		for _, term := range vocab {
			if !expected[term] {
				t.Errorf("Unexpected term in vocabulary: %q", term)
			}
		}
	})
}

func TestExpandTokensWithFuzzy(t *testing.T) {
	vocabulary := []string{
		"deploy", "deployment", "deployed",
		"config", "configuration", "configured",
		"server", "service", "servic",
		"user", "profil", "data",
	}

	t.Run("empty tokens returns nil", func(t *testing.T) {
		result := ExpandTokensWithFuzzy(nil, vocabulary)
		if result != nil {
			t.Errorf("Expected nil, got %v", result)
		}
	})

	t.Run("empty vocabulary returns nil", func(t *testing.T) {
		result := ExpandTokensWithFuzzy([]string{"test"}, nil)
		if result != nil {
			t.Errorf("Expected nil, got %v", result)
		}
	})

	t.Run("short tokens are skipped", func(t *testing.T) {
		result := ExpandTokensWithFuzzy([]string{"cat", "dog", "go"}, vocabulary)
		if len(result) != 0 {
			t.Errorf("Expected no matches for short tokens, got %v", result)
		}
	})

	t.Run("typo matches found", func(t *testing.T) {
		// "deplyo" is a typo for "deploy" (transposition)
		result := ExpandTokensWithFuzzy([]string{"deplyo"}, vocabulary)
		if len(result) == 0 {
			t.Error("Expected fuzzy matches for 'deplyo', got none")
			return
		}

		foundDeploy := false
		for _, m := range result {
			if m.Matched == "deploy" {
				foundDeploy = true
				if m.Original != "deplyo" {
					t.Errorf("Expected original 'deplyo', got %q", m.Original)
				}
				if m.Distance <= 0 {
					t.Errorf("Expected positive distance, got %d", m.Distance)
				}
			}
		}
		if !foundDeploy {
			t.Error("Expected 'deploy' in fuzzy matches for 'deplyo'")
		}
	})

	t.Run("very different term has no matches", func(t *testing.T) {
		result := ExpandTokensWithFuzzy([]string{"zzzzzzzzz"}, vocabulary)
		if len(result) != 0 {
			t.Errorf("Expected no matches for very different term, got %v", result)
		}
	})

	t.Run("exact match is excluded", func(t *testing.T) {
		result := ExpandTokensWithFuzzy([]string{"deploy"}, vocabulary)
		for _, m := range result {
			if m.Matched == "deploy" {
				t.Error("Exact match 'deploy' should be excluded from fuzzy matches")
			}
		}
	})
}

func TestMergeTokensWithFuzzy(t *testing.T) {
	t.Run("no fuzzy matches returns original tokens", func(t *testing.T) {
		tokens := []string{"hello", "world"}
		result := MergeTokensWithFuzzy(tokens, nil)
		if len(result) != 2 {
			t.Errorf("Expected 2 tokens, got %d", len(result))
		}
	})

	t.Run("fuzzy matches are appended", func(t *testing.T) {
		tokens := []string{"deplyo"}
		matches := []FuzzyMatch{
			{Original: "deplyo", Matched: "deploy", Distance: 2, Similarity: 0.5},
			{Original: "deplyo", Matched: "deployed", Distance: 2, Similarity: 0.4},
		}

		result := MergeTokensWithFuzzy(tokens, matches)
		if len(result) != 3 {
			t.Errorf("Expected 3 tokens, got %d: %v", len(result), result)
		}

		// Original should come first
		if result[0] != "deplyo" {
			t.Errorf("Expected first token to be 'deplyo', got %q", result[0])
		}
	})

	t.Run("duplicates are removed", func(t *testing.T) {
		tokens := []string{"deploy"}
		matches := []FuzzyMatch{
			{Original: "deploy", Matched: "deploy", Distance: 0, Similarity: 1.0},
		}

		result := MergeTokensWithFuzzy(tokens, matches)
		if len(result) != 1 {
			t.Errorf("Expected 1 token (deduped), got %d: %v", len(result), result)
		}
	})

	t.Run("original tokens take precedence", func(t *testing.T) {
		tokens := []string{"a", "b", "c"}
		matches := []FuzzyMatch{
			{Original: "a", Matched: "b", Distance: 1, Similarity: 0.5},
			{Original: "a", Matched: "d", Distance: 1, Similarity: 0.5},
		}

		result := MergeTokensWithFuzzy(tokens, matches)
		// a, b, c are original; d is new from fuzzy; b is deduped
		if len(result) != 4 {
			t.Errorf("Expected 4 tokens, got %d: %v", len(result), result)
		}
	})
}
