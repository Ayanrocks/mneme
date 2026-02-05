package index

import (
	"testing"
)

func TestIsStopword(t *testing.T) {
	tests := []struct {
		name     string
		token    string
		expected bool
	}{
		// Go keywords
		{"go func keyword", "func", true},
		{"go return keyword", "return", true},
		{"go package keyword", "package", true},
		{"go import keyword", "import", true},
		{"go defer keyword", "defer", true},
		{"go interface keyword", "interface", true},

		// Python keywords
		{"python def keyword", "def", true},
		{"python class keyword", "class", true},
		{"python lambda keyword", "lambda", true},

		// JavaScript keywords
		{"js function keyword", "function", true},
		{"js const keyword", "const", true},
		{"js let keyword", "let", true},

		// Java/C# keywords
		{"java public keyword", "public", true},
		{"java private keyword", "private", true},
		{"java static keyword", "static", true},

		// Common short tokens
		{"short token id", "id", true},
		{"short token err", "err", true},
		{"short token val", "val", true},

		// Non-stopwords (normal code identifiers)
		{"normal word user", "user", false},
		{"normal word config", "config", false},
		{"normal word handler", "handler", false},
		{"normal word service", "service", false},
		{"normal word database", "database", false},
		{"normal word request", "request", false},

		// Edge cases
		{"empty string", "", false},
		{"single char", "x", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsStopword(tt.token)
			if result != tt.expected {
				t.Errorf("IsStopword(%q) = %v, expected %v", tt.token, result, tt.expected)
			}
		})
	}
}

func TestFilterStopwords_Extended(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected []string
	}{
		{
			name:     "empty slice",
			input:    []string{},
			expected: []string{},
		},
		{
			name:     "no stopwords",
			input:    []string{"user", "profil", "config", "handler"},
			expected: []string{"user", "profil", "config", "handler"},
		},
		{
			name:     "all stopwords",
			input:    []string{"func", "return", "class", "public"},
			expected: []string{},
		},
		{
			name:     "mixed content",
			input:    []string{"get", "func", "user", "return", "profil"},
			expected: []string{"get", "user", "profil"},
		},
		{
			name:     "preserves order",
			input:    []string{"alpha", "func", "beta", "return", "gamma"},
			expected: []string{"alpha", "beta", "gamma"},
		},
		{
			name:     "handles duplicates",
			input:    []string{"user", "user", "func", "user"},
			expected: []string{"user", "user", "user"},
		},
		{
			name:     "programming noise filtered",
			input:    []string{"id", "user", "err", "handler", "val", "config"},
			expected: []string{"user", "handler", "config"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FilterStopwords(tt.input)
			if len(result) != len(tt.expected) {
				t.Errorf("FilterStopwords(%v) = %v (len=%d), expected %v (len=%d)",
					tt.input, result, len(result), tt.expected, len(tt.expected))
				return
			}
			for i, token := range result {
				if token != tt.expected[i] {
					t.Errorf("FilterStopwords()[%d] = %q, expected %q", i, token, tt.expected[i])
				}
			}
		})
	}
}

func TestProgrammingStopwordsMap(t *testing.T) {
	// Verify the stopwords map contains expected categories

	t.Run("contains Go keywords", func(t *testing.T) {
		goKeywords := []string{"func", "return", "defer", "go", "chan", "select"}
		for _, kw := range goKeywords {
			if !ProgrammingStopwords[kw] {
				t.Errorf("Expected %q to be in ProgrammingStopwords", kw)
			}
		}
	})

	t.Run("contains Python keywords", func(t *testing.T) {
		pyKeywords := []string{"def", "class", "lambda", "yield", "async", "await"}
		for _, kw := range pyKeywords {
			if !ProgrammingStopwords[kw] {
				t.Errorf("Expected %q to be in ProgrammingStopwords", kw)
			}
		}
	})

	t.Run("contains common noise tokens", func(t *testing.T) {
		noiseTokens := []string{"id", "err", "val", "arg", "ctx", "fmt"}
		for _, tok := range noiseTokens {
			if !ProgrammingStopwords[tok] {
				t.Errorf("Expected %q to be in ProgrammingStopwords", tok)
			}
		}
	})

	t.Run("map is not empty", func(t *testing.T) {
		if len(ProgrammingStopwords) < 50 {
			t.Errorf("ProgrammingStopwords should have many entries, got %d", len(ProgrammingStopwords))
		}
	})
}
