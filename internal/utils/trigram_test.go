package utils

import (
	"sort"
	"testing"
)

func TestGenerateTrigrams(t *testing.T) {
	tests := []struct {
		name     string
		term     string
		expected []string
	}{
		{
			name:     "empty string",
			term:     "",
			expected: nil,
		},
		{
			name:     "single character",
			term:     "a",
			expected: []string{"$$a", "$a$", "a$$"},
		},
		{
			name:     "two characters",
			term:     "ab",
			expected: []string{"$$a", "$ab", "ab$", "b$$"},
		},
		{
			name:     "three characters",
			term:     "cat",
			expected: []string{"$$c", "$ca", "cat", "at$", "t$$"},
		},
		{
			name:     "longer word",
			term:     "hello",
			expected: []string{"$$h", "$he", "hel", "ell", "llo", "lo$", "o$$"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GenerateTrigrams(tt.term)
			if len(result) != len(tt.expected) {
				t.Errorf("GenerateTrigrams(%q) = %v (len %d), expected %v (len %d)",
					tt.term, result, len(result), tt.expected, len(tt.expected))
				return
			}
			for i, tri := range result {
				if tri != tt.expected[i] {
					t.Errorf("GenerateTrigrams(%q)[%d] = %q, expected %q",
						tt.term, i, tri, tt.expected[i])
				}
			}
		})
	}
}

func TestTrigramSimilarity(t *testing.T) {
	tests := []struct {
		name   string
		a      string
		b      string
		minSim float64
		maxSim float64
	}{
		{
			name:   "identical strings",
			a:      "hello",
			b:      "hello",
			minSim: 1.0,
			maxSim: 1.0,
		},
		{
			name:   "completely different",
			a:      "abc",
			b:      "xyz",
			minSim: 0.0,
			maxSim: 0.15, // only padding trigrams might overlap
		},
		{
			name:   "one edit distance apart",
			a:      "hello",
			b:      "helo",
			minSim: 0.5,
			maxSim: 1.0,
		},
		{
			name:   "similar words",
			a:      "deploy",
			b:      "deplyo",
			minSim: 0.4,
			maxSim: 1.0,
		},
		{
			name:   "empty and non-empty",
			a:      "",
			b:      "hello",
			minSim: 0.0,
			maxSim: 0.0,
		},
		{
			name:   "both empty",
			a:      "",
			b:      "",
			minSim: 0.0,
			maxSim: 0.0, // no trigrams to compare
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sim := TrigramSimilarity(tt.a, tt.b)
			if sim < tt.minSim || sim > tt.maxSim {
				t.Errorf("TrigramSimilarity(%q, %q) = %v, expected in range [%v, %v]",
					tt.a, tt.b, sim, tt.minSim, tt.maxSim)
			}
		})
	}

	// Symmetry check
	t.Run("symmetry", func(t *testing.T) {
		sim1 := TrigramSimilarity("deployment", "deploymant")
		sim2 := TrigramSimilarity("deploymant", "deployment")
		if sim1 != sim2 {
			t.Errorf("TrigramSimilarity is not symmetric: %v != %v", sim1, sim2)
		}
	})
}

func TestBuildTrigramIndex(t *testing.T) {
	terms := []string{"cat", "car", "cart", "dog"}
	index := BuildTrigramIndex(terms)

	t.Run("non-empty index", func(t *testing.T) {
		if len(index) == 0 {
			t.Error("Expected non-empty trigram index")
		}
	})

	t.Run("shared trigrams map to multiple terms", func(t *testing.T) {
		// "cat" and "car" share "$ca" trigram
		terms, exists := index["$ca"]
		if !exists {
			t.Error("Expected $ca trigram in index")
			return
		}
		if len(terms) < 2 {
			t.Errorf("Expected at least 2 terms for $ca trigram, got %d", len(terms))
		}
	})

	t.Run("empty terms", func(t *testing.T) {
		emptyIndex := BuildTrigramIndex(nil)
		if len(emptyIndex) != 0 {
			t.Errorf("Expected empty index, got %d entries", len(emptyIndex))
		}
	})
}

func TestFindCandidates(t *testing.T) {
	vocabulary := []string{"deploy", "deployment", "deployed", "config", "configuration", "server", "service"}
	triIndex := BuildTrigramIndex(vocabulary)

	tests := []struct {
		name      string
		query     string
		threshold float64
		wantAny   bool
		wantTerms []string // if non-nil, expected results (subset check)
	}{
		{
			name:      "similar terms found",
			query:     "deplyo",
			threshold: 0.3,
			wantAny:   true,
			wantTerms: []string{"deploy"},
		},
		{
			name:      "exact match excluded",
			query:     "deploy",
			threshold: 0.3,
			wantAny:   true,
			// "deploy" itself should be excluded, but "deployed" and "deployment" should match
		},
		{
			name:      "very different term",
			query:     "zzzzzzz",
			threshold: 0.3,
			wantAny:   false,
		},
		{
			name:      "empty query",
			query:     "",
			threshold: 0.3,
			wantAny:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			candidates := FindCandidates(tt.query, triIndex, tt.threshold)
			if tt.wantAny && len(candidates) == 0 {
				t.Error("Expected at least one candidate, got none")
			}
			if !tt.wantAny && len(candidates) > 0 {
				t.Errorf("Expected no candidates, got %v", candidates)
			}
			if tt.wantTerms != nil {
				sort.Strings(candidates)
				for _, wanted := range tt.wantTerms {
					found := false
					for _, c := range candidates {
						if c == wanted {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("Expected candidate %q not found in %v", wanted, candidates)
					}
				}
			}
		})
	}
}
