package display

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
	"unicode"

	"mneme/internal/core"
	"mneme/internal/index"
	"mneme/internal/storage"

	"github.com/fatih/color"
)

const (
	// SnippetContextChars defines how many characters to show before/after the match
	SnippetContextChars = 50
	// MaxSnippetsPerResult limits the number of snippets shown per document
	MaxSnippetsPerResult = 3
	// MaxSnippetLength is the maximum length of a single snippet
	MaxSnippetLength = 150
)

var (
	pathColor      = color.New(color.FgCyan, color.Bold).SprintFunc()
	lineNumColor   = color.New(color.FgYellow).SprintFunc()
	matchColor     = color.New(color.FgRed, color.Bold).SprintFunc()
	scoreColor     = color.New(color.FgGreen).SprintFunc()
	separatorColor = color.New(color.FgWhite).SprintFunc()
)

// FormatSearchResult takes a document path and query tokens, reads the file,
// and returns a formatted SearchResult with snippets
func FormatSearchResult(docPath string, queryTokens []string, score float64) (*core.SearchResult, error) {
	lines, err := storage.ReadFileContents(docPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", docPath, err)
	}

	result := &core.SearchResult{
		DocPath:  docPath,
		Score:    score,
		Snippets: []core.Snippet{},
	}

	type matchLine struct {
		lineNum     int
		line        string
		matches     []core.HighlightRange
		score       int // score based on unique terms + total matches
		uniqueTerms int
	}

	var candidates []matchLine
	totalMatches := 0

	// Find all matching lines
	for lineNum, line := range lines {
		matches := findMatchesInLine(line, queryTokens)
		if len(matches) > 0 {
			totalMatches += len(matches)

			// Calculate score:
			// 1. Calculate unique terms matched in this line
			// 2. Score = (UniqueTerms * 100) + (TotalMatches)
			// This heavily prioritizes lines matching multiple different terms (e.g. "find" AND "token")
			// over lines matching the same term multiple times.

			matchedTerms := make(map[string]bool)
			lineLower := strings.ToLower(line)
			for _, token := range queryTokens {
				if len(token) > 0 && strings.Contains(lineLower, strings.ToLower(token)) {
					matchedTerms[token] = true
				}
			}
			uniqueCount := len(matchedTerms)
			lineScore := (uniqueCount * 100) + len(matches)

			candidates = append(candidates, matchLine{
				lineNum:     lineNum + 1,
				line:        line,
				matches:     matches,
				score:       lineScore,
				uniqueTerms: uniqueCount,
			})
		}
	}

	// Sort candidates by score descending
	// Use a stable sort to preserve line order for ties? No, we want score.
	// We can implement a simple sort.
	for i := 0; i < len(candidates)-1; i++ {
		for j := i + 1; j < len(candidates); j++ {
			if candidates[j].score > candidates[i].score {
				candidates[i], candidates[j] = candidates[j], candidates[i]
			}
		}
	}

	// Select top snippets, but try to respect original line order if scores are somewhat close?
	// For now, simpler is better: Top scores win.
	// But we might want to re-sort by line number for display?
	// Usually snippets are shown in file order.

	topCandidates := candidates
	if len(topCandidates) > MaxSnippetsPerResult {
		topCandidates = topCandidates[:MaxSnippetsPerResult]
	}

	// Re-sort selected snippets by line number for coherent display
	for i := 0; i < len(topCandidates)-1; i++ {
		for j := i + 1; j < len(topCandidates); j++ {
			if topCandidates[j].lineNum < topCandidates[i].lineNum {
				topCandidates[i], topCandidates[j] = topCandidates[j], topCandidates[i]
			}
		}
	}

	// Create snippets
	for _, c := range topCandidates {
		snippet := createSnippet(c.lineNum, c.line, c.matches)
		result.Snippets = append(result.Snippets, snippet)
	}

	result.MatchCount = totalMatches
	return result, nil
}

// findMatchesInLine finds all positions where query tokens match in a line
func findMatchesInLine(line string, queryTokens []string) []core.HighlightRange {
	var matches []core.HighlightRange
	lineLower := strings.ToLower(line)

	// Sort tokens by length (descending) to prioritize longer matches
	// This helps avoid highlighting "u" inside "unique" if "unique" is also a match
	sortedTokens := make([]string, len(queryTokens))
	copy(sortedTokens, queryTokens)
	sort.Slice(sortedTokens, func(i, j int) bool {
		return len(sortedTokens[i]) > len(sortedTokens[j])
	})

	// Track covered ranges to avoid overlapping or sub-matches
	coveredMask := make([]bool, len(line))

	for _, token := range sortedTokens {
		searchTermLower := strings.ToLower(token)
		if len(searchTermLower) == 0 {
			continue
		}

		startPos := 0
		for {
			idx := strings.Index(lineLower[startPos:], searchTermLower)
			if idx == -1 {
				break
			}

			actualStart := startPos + idx
			actualEnd := actualStart + len(token)

			// Bounds check: skip if match extends beyond the line
			if actualEnd > len(line) {
				startPos = actualStart + 1
				if startPos >= len(lineLower) {
					break
				}
				continue
			}

			// Check if this range is already covered
			isCovered := false
			for k := actualStart; k < actualEnd; k++ {
				if coveredMask[k] {
					isCovered = true
					break
				}
			}

			if !isCovered {
				// For short tokens (len < 4), enforce word boundaries
				// unless it's part of a larger matched token (already handled by sorting/masking)
				// But here we are looking for *this* token.
				isShort := len(token) < 4
				isValid := true

				if isShort {
					// Check word boundaries
					if !isWordBoundary(lineLower, actualStart, actualEnd) {
						isValid = false
					}
				}

				if isValid {
					matches = append(matches, core.HighlightRange{
						Start: actualStart,
						End:   actualEnd,
					})
					// Mark as covered
					for k := actualStart; k < actualEnd; k++ {
						coveredMask[k] = true
					}
				}
			}

			startPos = actualStart + 1
			if startPos >= len(lineLower) {
				break
			}
		}
	}

	// Merge overlapping ranges (fallback, though masking handles most)
	matches = mergeRanges(matches)
	return matches
}

// isWordBoundary checks if the match at start:end is a complete word
func isWordBoundary(line string, start, end int) bool {
	// Check start boundary
	if start > 0 {
		prevChar := rune(line[start-1])
		if unicode.IsLetter(prevChar) || unicode.IsDigit(prevChar) || prevChar == '_' {
			return false
		}
	}

	// Check end boundary
	if end < len(line) {
		nextChar := rune(line[end])
		if unicode.IsLetter(nextChar) || unicode.IsDigit(nextChar) || nextChar == '_' {
			return false
		}
	}

	return true
}

// mergeRanges merges overlapping highlight ranges
func mergeRanges(ranges []core.HighlightRange) []core.HighlightRange {
	if len(ranges) <= 1 {
		return ranges
	}

	// Sort by start position (simple bubble sort for small arrays)
	for i := 0; i < len(ranges)-1; i++ {
		for j := i + 1; j < len(ranges); j++ {
			if ranges[j].Start < ranges[i].Start {
				ranges[i], ranges[j] = ranges[j], ranges[i]
			}
		}
	}

	merged := []core.HighlightRange{ranges[0]}
	for i := 1; i < len(ranges); i++ {
		last := &merged[len(merged)-1]
		if ranges[i].Start <= last.End {
			if ranges[i].End > last.End {
				last.End = ranges[i].End
			}
		} else {
			merged = append(merged, ranges[i])
		}
	}

	return merged
}

// createSnippet creates a snippet with the matched content
func createSnippet(lineNum int, line string, matches []core.HighlightRange) core.Snippet {
	// Trim leading whitespace and track how much was removed
	trimmedLine := strings.TrimLeft(line, " \t")
	leadingTrimmed := len(line) - len(trimmedLine)

	// Adjust match positions for leading trim
	adjustedMatches := make([]core.HighlightRange, 0, len(matches))
	for _, m := range matches {
		newStart := m.Start - leadingTrimmed
		newEnd := m.End - leadingTrimmed
		if newStart >= 0 && newEnd <= len(trimmedLine) {
			adjustedMatches = append(adjustedMatches, core.HighlightRange{
				Start: newStart,
				End:   newEnd,
			})
		}
	}

	content := strings.TrimRight(trimmedLine, " \t")
	highlights := adjustedMatches

	// If line is still too long, create a focused snippet around the first match
	if len(content) > MaxSnippetLength && len(highlights) > 0 {
		firstMatch := highlights[0]
		start := firstMatch.Start - SnippetContextChars
		end := firstMatch.End + SnippetContextChars

		if start < 0 {
			start = 0
		}
		if end > len(content) {
			end = len(content)
		}

		// Adjust to word boundaries
		for start > 0 && !unicode.IsSpace(rune(content[start])) {
			start--
		}
		for end < len(content) && !unicode.IsSpace(rune(content[end-1])) {
			end++
		}
		if end > len(content) {
			end = len(content)
		}

		prefix := ""
		suffix := ""
		if start > 0 {
			prefix = "..."
			// Skip leading whitespace after prefix
			for start < len(content) && unicode.IsSpace(rune(content[start])) {
				start++
			}
		}
		if end < len(content) {
			suffix = "..."
		}

		snippetContent := content[start:end]
		content = prefix + snippetContent + suffix

		// Adjust highlight positions for the new content
		newHighlights := []core.HighlightRange{}
		for _, m := range highlights {
			if m.Start >= start && m.End <= end {
				newHighlights = append(newHighlights, core.HighlightRange{
					Start: m.Start - start + len(prefix),
					End:   m.End - start + len(prefix),
				})
			}
		}
		highlights = newHighlights
	}

	return core.Snippet{
		LineNumber: lineNum,
		Content:    content,
		Highlights: highlights,
	}
}

// PrintResult prints a formatted search result to stdout
func PrintResult(result *core.SearchResult, showScore bool) {
	// Print document path
	fmt.Printf("%s\n", pathColor(result.DocPath))

	// Print score if requested
	if showScore && result.Score > 0 {
		fmt.Printf("  Score: %s\n", scoreColor(fmt.Sprintf("%.4f", result.Score)))
	}

	// Print snippets
	for _, snippet := range result.Snippets {
		linePrefix := fmt.Sprintf("  %s: ", lineNumColor(fmt.Sprintf("Ln %d", snippet.LineNumber)))
		fmt.Print(linePrefix)

		// Print content with highlights
		printHighlightedContent(snippet.Content, snippet.Highlights)
		fmt.Println()
	}

	// Print separator
	fmt.Println()
}

// printHighlightedContent prints the content with highlighted matches
func printHighlightedContent(content string, highlights []core.HighlightRange) {
	lastEnd := 0

	for _, h := range highlights {
		// Bounds checking
		if h.Start < 0 || h.End > len(content) || h.Start >= h.End {
			continue
		}
		if h.Start < lastEnd {
			continue
		}

		// Print content before highlight
		if h.Start > lastEnd {
			fmt.Print(content[lastEnd:h.Start])
		}

		// Print highlighted match
		fmt.Print(matchColor(content[h.Start:h.End]))
		lastEnd = h.End
	}

	// Print remaining content
	if lastEnd < len(content) {
		fmt.Print(content[lastEnd:])
	}
}

// PrintResults prints multiple search results
func PrintResults(results []*core.SearchResult, showScore bool, query string) {
	if len(results) == 0 {
		fmt.Printf("No results found for: %s\n", query)
		return
	}

	// Print header
	fmt.Printf("\n%s\n", separatorColor(strings.Repeat("─", 60)))
	fmt.Printf("Found %s results for: %s\n",
		scoreColor(fmt.Sprintf("%d", len(results))),
		matchColor(query))
	fmt.Printf("%s\n\n", separatorColor(strings.Repeat("─", 60)))

	for i, result := range results {
		fmt.Printf("%s. ", lineNumColor(fmt.Sprintf("%d", i+1)))
		PrintResult(result, showScore)
	}
}

// PrintSimpleResult prints a condensed result format (path + first snippet)
func PrintSimpleResult(docPath string, queryTokens []string) {
	result, err := FormatSearchResult(docPath, queryTokens, 0)
	if err != nil {
		// Fallback to just printing the path
		fmt.Printf("%s\n", pathColor(docPath))
		return
	}

	fmt.Printf("%s", pathColor(docPath))
	if len(result.Snippets) > 0 {
		fmt.Printf(" — %s", truncateSnippet(result.Snippets[0].Content, 80))
	}
	fmt.Println()
}

// truncateSnippet truncates a snippet to the specified length
func truncateSnippet(content string, maxLen int) string {
	// Remove excess whitespace
	content = regexp.MustCompile(`\s+`).ReplaceAllString(content, " ")
	content = strings.TrimSpace(content)

	if len(content) <= maxLen {
		return content
	}

	return content[:maxLen-3] + "..."
}

// CreateSearchResultFromTokens creates a SearchResult by re-tokenizing the query for matching
func CreateSearchResultFromTokens(docPath string, originalQuery string, score float64) (*core.SearchResult, error) {
	// Tokenize the original query to get search terms
	queryTokens := index.TokenizeQuery(originalQuery)
	return FormatSearchResult(docPath, queryTokens, score)
}
