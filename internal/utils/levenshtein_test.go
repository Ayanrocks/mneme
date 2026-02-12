package utils

import (
	"testing"
)

func TestLevenshteinDistance(t *testing.T) {
	tests := []struct {
		name     string
		a        string
		b        string
		expected int
	}{
		{
			name:     "identical strings",
			a:        "hello",
			b:        "hello",
			expected: 0,
		},
		{
			name:     "single insertion",
			a:        "cat",
			b:        "cats",
			expected: 1,
		},
		{
			name:     "single deletion",
			a:        "cats",
			b:        "cat",
			expected: 1,
		},
		{
			name:     "single substitution",
			a:        "cat",
			b:        "car",
			expected: 1,
		},
		{
			name:     "transposition (counts as 2)",
			a:        "ab",
			b:        "ba",
			expected: 2,
		},
		{
			name:     "empty to non-empty",
			a:        "",
			b:        "hello",
			expected: 5,
		},
		{
			name:     "non-empty to empty",
			a:        "hello",
			b:        "",
			expected: 5,
		},
		{
			name:     "both empty",
			a:        "",
			b:        "",
			expected: 0,
		},
		{
			name:     "completely different",
			a:        "abc",
			b:        "xyz",
			expected: 3,
		},
		{
			name:     "deployment typo",
			a:        "deployment",
			b:        "deploymant",
			expected: 1,
		},
		{
			name:     "kubernetes typo",
			a:        "kubernetes",
			b:        "kuberntes",
			expected: 1,
		},
		{
			name:     "two edits apart",
			a:        "kitten",
			b:        "sitting",
			expected: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := LevenshteinDistance(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("LevenshteinDistance(%q, %q) = %d, expected %d",
					tt.a, tt.b, result, tt.expected)
			}
		})
	}

	// Symmetry check
	t.Run("symmetry", func(t *testing.T) {
		d1 := LevenshteinDistance("deploy", "deplyo")
		d2 := LevenshteinDistance("deplyo", "deploy")
		if d1 != d2 {
			t.Errorf("LevenshteinDistance is not symmetric: %d != %d", d1, d2)
		}
	})
}

func TestIsWithinEditDistance(t *testing.T) {
	tests := []struct {
		name     string
		a        string
		b        string
		maxDist  int
		expected bool
	}{
		{
			name:     "identical within 0",
			a:        "hello",
			b:        "hello",
			maxDist:  0,
			expected: true,
		},
		{
			name:     "one edit within 1",
			a:        "cat",
			b:        "car",
			maxDist:  1,
			expected: true,
		},
		{
			name:     "one edit within 2",
			a:        "cat",
			b:        "car",
			maxDist:  2,
			expected: true,
		},
		{
			name:     "three edits not within 2",
			a:        "abc",
			b:        "xyz",
			maxDist:  2,
			expected: false,
		},
		{
			name:     "length difference exceeds max — early exit",
			a:        "hi",
			b:        "hello",
			maxDist:  2,
			expected: false,
		},
		{
			name:     "deployment typo within 2",
			a:        "deployment",
			b:        "deploymant",
			maxDist:  2,
			expected: true,
		},
		{
			name:     "empty strings within 0",
			a:        "",
			b:        "",
			maxDist:  0,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsWithinEditDistance(tt.a, tt.b, tt.maxDist)
			if result != tt.expected {
				t.Errorf("IsWithinEditDistance(%q, %q, %d) = %v, expected %v",
					tt.a, tt.b, tt.maxDist, result, tt.expected)
			}
		})
	}
}

func TestMin3(t *testing.T) {
	tests := []struct {
		a, b, c  int
		expected int
	}{
		{1, 2, 3, 1},
		{3, 1, 2, 1},
		{2, 3, 1, 1},
		{5, 5, 5, 5},
		{0, 1, 2, 0},
	}

	for _, tt := range tests {
		result := min3(tt.a, tt.b, tt.c)
		if result != tt.expected {
			t.Errorf("min3(%d, %d, %d) = %d, expected %d",
				tt.a, tt.b, tt.c, result, tt.expected)
		}
	}
}
