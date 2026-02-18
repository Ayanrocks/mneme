package utils

// LevenshteinDistance calculates the minimum number of single-character edits
// (insertions, deletions, substitutions) needed to transform string a into string b.
// Uses the Wagner-Fischer algorithm with single-row space optimization: O(min(m,n)) space.
func LevenshteinDistance(a, b string) int {
	runesA := []rune(a)
	runesB := []rune(b)

	// Ensure a is the shorter string for space optimization
	if len(runesA) > len(runesB) {
		runesA, runesB = runesB, runesA
	}

	m := len(runesA)
	n := len(runesB)

	if m == 0 {
		return n
	}

	// Single row of the DP matrix
	prev := make([]int, m+1)
	for i := 0; i <= m; i++ {
		prev[i] = i
	}

	for j := 1; j <= n; j++ {
		curr := make([]int, m+1)
		curr[0] = j

		for i := 1; i <= m; i++ {
			cost := 1
			if runesA[i-1] == runesB[j-1] {
				cost = 0
			}

			// Minimum of insert, delete, substitute
			del := prev[i] + 1
			ins := curr[i-1] + 1
			sub := prev[i-1] + cost

			curr[i] = min3(del, ins, sub)
		}

		prev = curr
	}

	return prev[m]
}

// IsWithinEditDistance checks if two strings are within the given maximum
// edit distance. Includes an early exit optimization based on length difference.
func IsWithinEditDistance(a, b string, maxDist int) bool {
	// Quick length-difference check — if lengths differ by more than
	// maxDist, edit distance must exceed it
	lenDiff := len([]rune(a)) - len([]rune(b))
	if lenDiff < 0 {
		lenDiff = -lenDiff
	}
	if lenDiff > maxDist {
		return false
	}

	return LevenshteinDistance(a, b) <= maxDist
}

// DamerauLevenshteinDistance calculates the optimal string alignment distance
// between two strings. Unlike standard Levenshtein, this counts transpositions
// of two adjacent characters as a single edit operation (distance 1).
// This is the restricted edit distance variant (no substring may be edited more than once).
func DamerauLevenshteinDistance(a, b string) int {
	runesA := []rune(a)
	runesB := []rune(b)

	m := len(runesA)
	n := len(runesB)

	if m == 0 {
		return n
	}
	if n == 0 {
		return m
	}

	// We need two previous rows for the transposition check
	prevPrev := make([]int, n+1)
	prev := make([]int, n+1)
	curr := make([]int, n+1)

	for j := 0; j <= n; j++ {
		prev[j] = j
	}

	for i := 1; i <= m; i++ {
		curr[0] = i

		for j := 1; j <= n; j++ {
			cost := 1
			if runesA[i-1] == runesB[j-1] {
				cost = 0
			}

			// Minimum of insert, delete, substitute
			del := prev[j] + 1
			ins := curr[j-1] + 1
			sub := prev[j-1] + cost

			curr[j] = min3(del, ins, sub)

			// Check for transposition
			if i > 1 && j > 1 &&
				runesA[i-1] == runesB[j-2] &&
				runesA[i-2] == runesB[j-1] {
				trans := prevPrev[j-2] + cost
				if trans < curr[j] {
					curr[j] = trans
				}
			}
		}

		// Rotate rows
		prevPrev, prev, curr = prev, curr, prevPrev
	}

	return prev[n]
}

// IsWithinDamerauDistance checks if two strings are within the given maximum
// Damerau-Levenshtein distance. Includes an early exit optimization based on
// length difference.
func IsWithinDamerauDistance(a, b string, maxDist int) bool {
	lenDiff := len([]rune(a)) - len([]rune(b))
	if lenDiff < 0 {
		lenDiff = -lenDiff
	}
	if lenDiff > maxDist {
		return false
	}

	return DamerauLevenshteinDistance(a, b) <= maxDist
}

// min3 returns the minimum of three integers.
func min3(a, b, c int) int {
	if a < b {
		if a < c {
			return a
		}
		return c
	}
	if b < c {
		return b
	}
	return c
}
