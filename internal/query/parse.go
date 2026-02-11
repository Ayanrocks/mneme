package query

import (
	"mneme/internal/core"
	"mneme/internal/index"
	"mneme/internal/logger"
	"strings"
)

// ParseQuery tokenizes a query string into stemmed tokens for BM25/VSM scoring.
// This is a convenience wrapper that maintains backward compatibility.
func ParseQuery(queryString string) []string {
	logger.Info("Query string: " + queryString)

	// Use the same tokenization pipeline as indexing for BM25 consistency
	tokens := index.TokenizeQuery(queryString)

	return tokens
}

// ParseQueryInput splits raw user input into two outputs:
//   - searchTerms: quoted phrases stay intact, unquoted words are kept individually (for snippet matching)
//   - stemmedTokens: all words individually stemmed via TokenizeQuery (for BM25/VSM scoring)
//
// Examples:
//
//	`"aws region"` → searchTerms=["aws region"], stemmedTokens=[stem("aws"), stem("region")]
//	`deploy production` → searchTerms=["deploy", "production"], stemmedTokens=[stem("deploy"), stem("production")]
//	`"error handling" golang` → searchTerms=["error handling", "golang"], stemmedTokens=[stem("error"), stem("handling"), stem("golang")]
func ParseQueryInput(raw string) (searchTerms []string, stemmedTokens []string) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}

	searchTerms = extractSearchTerms(raw)

	// Build stemmed tokens from all individual words across all terms
	var allWords []string
	for _, term := range searchTerms {
		words := strings.Fields(term)
		allWords = append(allWords, words...)
	}

	stemmedTokens = index.TokenizeQuery(strings.Join(allWords, " "))

	logger.Info("Query input parsed: searchTerms=" + strings.Join(searchTerms, ", ") +
		" stemmedTokens=" + strings.Join(stemmedTokens, ", "))

	return searchTerms, stemmedTokens
}

// extractSearchTerms parses raw input, keeping quoted substrings as single entries
// and splitting unquoted text by whitespace.
func extractSearchTerms(raw string) []string {
	var terms []string
	var current strings.Builder
	inQuotes := false

	for i := 0; i < len(raw); i++ {
		ch := raw[i]

		if ch == '"' {
			if inQuotes {
				// Closing quote — flush the phrase
				phrase := strings.TrimSpace(current.String())
				if phrase != "" {
					terms = append(terms, phrase)
				}
				current.Reset()
				inQuotes = false
			} else {
				// Opening quote — flush any accumulated unquoted words first
				unquoted := strings.TrimSpace(current.String())
				if unquoted != "" {
					words := strings.Fields(unquoted)
					terms = append(terms, words...)
				}
				current.Reset()
				inQuotes = true
			}
		} else {
			current.WriteByte(ch)
		}
	}

	// Flush remaining content
	remaining := strings.TrimSpace(current.String())
	if remaining != "" {
		if inQuotes {
			// Unclosed quote — treat the content as a phrase anyway
			terms = append(terms, remaining)
		} else {
			words := strings.Fields(remaining)
			terms = append(terms, words...)
		}
	}

	return terms
}

// FindQueryToken finds documents matching the given tokens using BM25+VSM ranking.
func FindQueryToken(segment *core.Segment, tokens []string) []string {
	logger.Info("Finding query token in segments using BM25+VSM ranking")

	// Use the ranking system that combines BM25 and VSM scores
	// Limit results to top 20 documents for performance
	return GetTopDocumentPaths(segment, tokens, MaxResults)
}
