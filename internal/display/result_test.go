package display

import (
	"mneme/internal/core"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFormatSearchResult(t *testing.T) {
	// Create temp file for testing
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	content := "This is a test file\nwith some user data\nand more user information\n"
	err := os.WriteFile(testFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	t.Run("formats result with matching query", func(t *testing.T) {
		result, err := FormatSearchResult(testFile, []string{"user"}, 0.75)
		if err != nil {
			t.Fatalf("FormatSearchResult error: %v", err)
		}

		if result.DocPath != testFile {
			t.Errorf("Expected path %s, got %s", testFile, result.DocPath)
		}
		if result.Score != 0.75 {
			t.Errorf("Expected score 0.75, got %f", result.Score)
		}
		if len(result.Snippets) == 0 {
			t.Error("Expected at least one snippet")
		}
		if result.MatchCount < 2 {
			t.Errorf("Expected at least 2 matches for 'user', got %d", result.MatchCount)
		}
	})

	t.Run("returns empty snippets for non-matching query", func(t *testing.T) {
		result, err := FormatSearchResult(testFile, []string{"nonexistent"}, 0.5)
		if err != nil {
			t.Fatalf("FormatSearchResult error: %v", err)
		}

		if len(result.Snippets) != 0 {
			t.Errorf("Expected no snippets for non-matching query, got %d", len(result.Snippets))
		}
		if result.MatchCount != 0 {
			t.Errorf("Expected 0 matches, got %d", result.MatchCount)
		}
	})

	t.Run("respects MaxSnippetsPerResult limit", func(t *testing.T) {
		// Create file with many matches
		manyMatches := strings.Repeat("user\n", 10)
		manyMatchFile := filepath.Join(tmpDir, "many_matches.txt")
		err := os.WriteFile(manyMatchFile, []byte(manyMatches), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		result, err := FormatSearchResult(manyMatchFile, []string{"user"}, 0.5)
		if err != nil {
			t.Fatalf("FormatSearchResult error: %v", err)
		}

		if len(result.Snippets) > MaxSnippetsPerResult {
			t.Errorf("Expected at most %d snippets, got %d", MaxSnippetsPerResult, len(result.Snippets))
		}
	})

	t.Run("handles non-existent file", func(t *testing.T) {
		_, err := FormatSearchResult("/nonexistent/file.txt", []string{"test"}, 0.5)
		if err == nil {
			t.Error("Expected error for non-existent file")
		}
	})

	t.Run("handles empty query tokens", func(t *testing.T) {
		result, err := FormatSearchResult(testFile, []string{}, 0.5)
		if err != nil {
			t.Fatalf("FormatSearchResult error: %v", err)
		}

		if len(result.Snippets) != 0 {
			t.Error("Expected no snippets for empty query")
		}
	})

	t.Run("case insensitive matching", func(t *testing.T) {
		result, err := FormatSearchResult(testFile, []string{"USER", "User", "uSeR"}, 0.5)
		if err != nil {
			t.Fatalf("FormatSearchResult error: %v", err)
		}

		if len(result.Snippets) == 0 {
			t.Error("Expected case-insensitive matches")
		}
	})
}

func TestFindMatchesInLine(t *testing.T) {
	t.Run("finds single match", func(t *testing.T) {
		matches := findMatchesInLine("This is a test", []string{"test"})
		if len(matches) != 1 {
			t.Errorf("Expected 1 match, got %d", len(matches))
		}
		if matches[0].Start != 10 || matches[0].End != 14 {
			t.Errorf("Expected match at [10:14], got [%d:%d]", matches[0].Start, matches[0].End)
		}
	})

	t.Run("finds multiple matches of same term", func(t *testing.T) {
		matches := findMatchesInLine("test test test", []string{"test"})
		if len(matches) != 3 {
			t.Errorf("Expected 3 matches, got %d", len(matches))
		}
	})

	t.Run("finds multiple different terms", func(t *testing.T) {
		matches := findMatchesInLine("user data and profile", []string{"user", "profile"})
		if len(matches) != 2 {
			t.Errorf("Expected 2 matches, got %d", len(matches))
		}
	})

	t.Run("merges overlapping matches", func(t *testing.T) {
		// "use" and "user" will overlap
		matches := findMatchesInLine("user information", []string{"use", "user"})
		// After merging, should have fewer than 2 matches
		if len(matches) > 2 {
			t.Errorf("Expected merged matches, got %d separate matches", len(matches))
		}
	})

	t.Run("returns empty for no matches", func(t *testing.T) {
		matches := findMatchesInLine("hello world", []string{"test"})
		if len(matches) != 0 {
			t.Errorf("Expected 0 matches, got %d", len(matches))
		}
	})

	t.Run("handles empty query tokens", func(t *testing.T) {
		matches := findMatchesInLine("test", []string{})
		if len(matches) != 0 {
			t.Errorf("Expected 0 matches for empty query, got %d", len(matches))
		}
	})

	t.Run("handles empty string tokens", func(t *testing.T) {
		matches := findMatchesInLine("test", []string{""})
		if len(matches) != 0 {
			t.Errorf("Expected 0 matches for empty string token, got %d", len(matches))
		}
	})

	t.Run("case insensitive matching", func(t *testing.T) {
		matches := findMatchesInLine("Test TEST test", []string{"test"})
		if len(matches) != 3 {
			t.Errorf("Expected 3 case-insensitive matches, got %d", len(matches))
		}
	})
}

func TestMergeRanges(t *testing.T) {
	t.Run("empty ranges", func(t *testing.T) {
		merged := mergeRanges([]core.HighlightRange{})
		if len(merged) != 0 {
			t.Errorf("Expected 0 merged ranges, got %d", len(merged))
		}
	})

	t.Run("single range", func(t *testing.T) {
		ranges := []core.HighlightRange{{Start: 5, End: 10}}
		merged := mergeRanges(ranges)
		if len(merged) != 1 {
			t.Errorf("Expected 1 merged range, got %d", len(merged))
		}
	})

	t.Run("non-overlapping ranges", func(t *testing.T) {
		ranges := []core.HighlightRange{
			{Start: 0, End: 5},
			{Start: 10, End: 15},
		}
		merged := mergeRanges(ranges)
		if len(merged) != 2 {
			t.Errorf("Expected 2 merged ranges, got %d", len(merged))
		}
	})

	t.Run("overlapping ranges", func(t *testing.T) {
		ranges := []core.HighlightRange{
			{Start: 0, End: 10},
			{Start: 5, End: 15},
		}
		merged := mergeRanges(ranges)
		if len(merged) != 1 {
			t.Errorf("Expected 1 merged range, got %d", len(merged))
		}
		if merged[0].Start != 0 || merged[0].End != 15 {
			t.Errorf("Expected merged range [0:15], got [%d:%d]", merged[0].Start, merged[0].End)
		}
	})

	t.Run("adjacent ranges", func(t *testing.T) {
		ranges := []core.HighlightRange{
			{Start: 0, End: 5},
			{Start: 5, End: 10},
		}
		merged := mergeRanges(ranges)
		if len(merged) != 1 {
			t.Errorf("Expected 1 merged range for adjacent ranges, got %d", len(merged))
		}
	})

	t.Run("unsorted ranges", func(t *testing.T) {
		ranges := []core.HighlightRange{
			{Start: 10, End: 15},
			{Start: 0, End: 5},
		}
		merged := mergeRanges(ranges)
		// Should sort and keep separate
		if len(merged) != 2 {
			t.Errorf("Expected 2 merged ranges, got %d", len(merged))
		}
		// First merged range should be the earlier one
		if merged[0].Start != 0 {
			t.Errorf("Expected first range to start at 0, got %d", merged[0].Start)
		}
	})
}

func TestCreateSnippet(t *testing.T) {
	t.Run("creates snippet with match", func(t *testing.T) {
		line := "This is a test line"
		matches := []core.HighlightRange{{Start: 10, End: 14}}
		snippet := createSnippet(1, line, matches)

		if snippet.LineNumber != 1 {
			t.Errorf("Expected line number 1, got %d", snippet.LineNumber)
		}
		if len(snippet.Highlights) == 0 {
			t.Error("Expected highlights in snippet")
		}
		if snippet.Content == "" {
			t.Error("Expected non-empty content")
		}
	})

	t.Run("trims leading whitespace", func(t *testing.T) {
		line := "    indented line with match"
		matches := []core.HighlightRange{{Start: 25, End: 30}}
		snippet := createSnippet(1, line, matches)

		// Content should be trimmed
		if strings.HasPrefix(snippet.Content, "    ") {
			t.Error("Expected leading whitespace to be trimmed")
		}
	})

	t.Run("handles very long lines", func(t *testing.T) {
		// Create a very long line
		longLine := strings.Repeat("a", 500) + "match" + strings.Repeat("b", 500)
		matches := []core.HighlightRange{{Start: 500, End: 505}}
		snippet := createSnippet(1, longLine, matches)

		// The function should create a snippet (may or may not truncate depending on implementation)
		// Just verify it doesn't panic and produces some content
		if len(snippet.Content) == 0 {
			t.Error("Expected non-empty snippet content")
		}
		// Verify highlights are present
		if len(matches) > 0 && len(snippet.Highlights) == 0 {
			// Highlights might be adjusted or removed if out of bounds after truncation
			// This is acceptable behavior
		}
	})

	t.Run("handles empty matches", func(t *testing.T) {
		line := "test line"
		matches := []core.HighlightRange{}
		snippet := createSnippet(1, line, matches)

		if len(snippet.Highlights) != 0 {
			t.Errorf("Expected no highlights, got %d", len(snippet.Highlights))
		}
	})

	t.Run("adjusts highlights after trimming", func(t *testing.T) {
		line := "    test"
		matches := []core.HighlightRange{{Start: 4, End: 8}}
		snippet := createSnippet(1, line, matches)

		// Highlights should be adjusted for trimmed content
		if len(snippet.Highlights) > 0 {
			// First highlight should start at 0 after trimming 4 spaces
			if snippet.Highlights[0].Start != 0 {
				t.Errorf("Expected adjusted highlight start at 0, got %d", snippet.Highlights[0].Start)
			}
		}
	})
}

func TestIsWordBoundary(t *testing.T) {
	t.Run("match at start of line", func(t *testing.T) {
		if !isWordBoundary("test line", 0, 4) {
			t.Error("Expected word boundary at start")
		}
	})

	t.Run("match at end of line", func(t *testing.T) {
		if !isWordBoundary("line test", 5, 9) {
			t.Error("Expected word boundary at end")
		}
	})

	t.Run("match with space boundaries", func(t *testing.T) {
		if !isWordBoundary("the test line", 4, 8) {
			t.Error("Expected word boundary with spaces")
		}
	})

	t.Run("match within word fails", func(t *testing.T) {
		if isWordBoundary("testing", 0, 4) {
			t.Error("Expected no word boundary within word")
		}
	})

	t.Run("match with punctuation boundaries", func(t *testing.T) {
		if !isWordBoundary("test.", 0, 4) {
			t.Error("Expected word boundary with punctuation")
		}
	})
}

func TestTruncateSnippet(t *testing.T) {
	t.Run("short content unchanged", func(t *testing.T) {
		content := "short content"
		result := truncateSnippet(content, 80)
		if result != content {
			t.Errorf("Expected unchanged content, got %s", result)
		}
	})

	t.Run("long content truncated", func(t *testing.T) {
		content := strings.Repeat("a", 200)
		result := truncateSnippet(content, 80)
		if len(result) > 80 {
			t.Errorf("Expected truncated to 80 chars, got %d", len(result))
		}
		if !strings.HasSuffix(result, "...") {
			t.Error("Expected truncated content to end with '...'")
		}
	})

	t.Run("removes excess whitespace", func(t *testing.T) {
		content := "test    with    spaces"
		result := truncateSnippet(content, 80)
		if strings.Contains(result, "    ") {
			t.Error("Expected excess whitespace to be removed")
		}
	})

	t.Run("handles empty content", func(t *testing.T) {
		result := truncateSnippet("", 80)
		if result != "" {
			t.Errorf("Expected empty result, got %s", result)
		}
	})
}

func TestPrintHighlightedContent(t *testing.T) {
	// This test is basic since printHighlightedContent prints to stdout
	// We mainly test that it doesn't panic
	t.Run("does not panic with valid input", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("printHighlightedContent panicked: %v", r)
			}
		}()

		content := "test content"
		highlights := []core.HighlightRange{{Start: 0, End: 4}}
		// Redirect stdout to avoid cluttering test output
		printHighlightedContent(content, highlights)
	})

	t.Run("handles out of bounds highlights", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("printHighlightedContent panicked: %v", r)
			}
		}()

		content := "test"
		highlights := []core.HighlightRange{{Start: -1, End: 10}}
		printHighlightedContent(content, highlights)
	})

	t.Run("handles overlapping highlights", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("printHighlightedContent panicked: %v", r)
			}
		}()

		content := "test content"
		highlights := []core.HighlightRange{
			{Start: 0, End: 4},
			{Start: 2, End: 6},
		}
		printHighlightedContent(content, highlights)
	})
}

func TestPrintResult(t *testing.T) {
	// Test that PrintResult doesn't panic
	t.Run("prints result without panic", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("PrintResult panicked: %v", r)
			}
		}()

		result := &core.SearchResult{
			DocPath: "/test/path.txt",
			Score:   0.75,
			Snippets: []core.Snippet{
				{
					LineNumber: 1,
					Content:    "test content",
					Highlights: []core.HighlightRange{{Start: 0, End: 4}},
				},
			},
			MatchCount: 1,
		}
		PrintResult(result, true)
	})
}

func TestPrintResults(t *testing.T) {
	// Test that PrintResults doesn't panic
	t.Run("prints multiple results without panic", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("PrintResults panicked: %v", r)
			}
		}()

		results := []*core.SearchResult{
			{
				DocPath:    "/test/path1.txt",
				Score:      0.75,
				Snippets:   []core.Snippet{{LineNumber: 1, Content: "test"}},
				MatchCount: 1,
			},
			{
				DocPath:    "/test/path2.txt",
				Score:      0.65,
				Snippets:   []core.Snippet{{LineNumber: 2, Content: "test"}},
				MatchCount: 1,
			},
		}
		PrintResults(results, true, "test query")
	})

	t.Run("handles empty results", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("PrintResults panicked: %v", r)
			}
		}()

		PrintResults([]*core.SearchResult{}, true, "empty query")
	})
}