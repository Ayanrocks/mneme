package index

import (
	"encoding/json"
	"strings"
	"sync"
	"unicode"

	"github.com/caneroj1/stemmer"
	"github.com/fatih/camelcase"
	"github.com/go-ego/gse"
)

var (
	// gseSegmenter is the global segmenter instance for CJK text
	gseSegmenter gse.Segmenter
	gseOnce      sync.Once
	gseInitErr   error
)

// initGSE initializes the gse segmenter (lazy, thread-safe)
func initGSE() error {
	gseOnce.Do(func() {
		// Load embedded dictionary - this works for Chinese/Japanese
		gseInitErr = gseSegmenter.LoadDictEmbed()
	})
	return gseInitErr
}

// containsCJK checks if the content contains CJK characters
func containsCJK(content string) bool {
	for _, r := range content {
		// Check for CJK Unified Ideographs and related ranges
		if unicode.Is(unicode.Han, r) || // Chinese
			unicode.Is(unicode.Hiragana, r) || // Japanese
			unicode.Is(unicode.Katakana, r) || // Japanese
			unicode.Is(unicode.Hangul, r) { // Korean
			return true
		}
	}
	return false
}

// TokenizeContent tokenizes content for BM25 indexing.
// It handles code files with camelCase, snake_case, and kebab-case identifiers.
// For CJK content, it uses gse segmentation.
// Returns a slice of normalized, stemmed, stopword-filtered tokens.
func TokenizeContent(content string) []string {
	// Check for binary content first
	if IsBinaryContent(content) {
		return []string{}
	}

	var tokens []string

	// Check if content contains CJK characters
	if containsCJK(content) {
		tokens = tokenizeMixed(content)
	} else {
		tokens = tokenizeCode(content)
	}

	// Filter out programming stopwords
	return FilterStopwords(tokens)
}

// tokenizeMixed handles content that may mix CJK and Latin text
func tokenizeMixed(content string) []string {
	var tokens []string

	// Initialize gse if needed
	if err := initGSE(); err != nil {
		// Fall back to code tokenization if gse fails
		return tokenizeCode(content)
	}

	// Use gse to segment the content
	segments := gseSegmenter.Segment([]byte(content))

	for _, seg := range segments {
		word := seg.Token().Text()

		// Skip empty or whitespace-only
		if strings.TrimSpace(word) == "" {
			continue
		}

		// For non-CJK words (like English identifiers), process as code
		if !containsCJK(word) {
			tokens = append(tokens, processIdentifier(word)...)
		} else {
			// For CJK words, just lowercase (no stemming for CJK)
			token := strings.ToLower(word)
			if len(token) > 0 {
				tokens = append(tokens, token)
			}
		}
	}

	return tokens
}

// tokenizeCode handles code/programming content with identifier splitting
func tokenizeCode(content string) []string {
	var tokens []string

	// Extract words/identifiers - matches sequences of letters, digits, underscores
	var currentWord strings.Builder
	for _, r := range content {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' {
			currentWord.WriteRune(r)
		} else {
			if currentWord.Len() > 0 {
				tokens = append(tokens, processIdentifier(currentWord.String())...)
				currentWord.Reset()
			}
		}
	}
	// Don't forget the last word
	if currentWord.Len() > 0 {
		tokens = append(tokens, processIdentifier(currentWord.String())...)
	}

	return tokens
}

// processIdentifier splits an identifier into tokens and applies normalization.
// Handles camelCase, PascalCase, snake_case, and mixed identifiers.
func processIdentifier(identifier string) []string {
	var result []string

	// First, split by underscores (snake_case)
	parts := strings.Split(identifier, "_")

	for _, part := range parts {
		if part == "" {
			continue
		}

		// Split camelCase/PascalCase using fatih/camelcase
		camelParts := camelcase.Split(part)

		for _, word := range camelParts {
			// Normalize: lowercase
			token := strings.ToLower(word)

			// Skip empty tokens
			if token == "" {
				continue
			}

			// Skip purely numeric tokens
			if isNumeric(token) {
				continue
			}

			// Skip very short tokens (single characters)
			if len(token) < 2 {
				continue
			}

			// Apply Porter stemming for BM25 consistency (only for Latin text)
			// Note: stemmer.Stem returns uppercase, so we lowercase after
			stemmed := strings.ToLower(stemmer.Stem(token))

			// Ensure stemmed result is still valid
			if stemmed != "" && len(stemmed) >= 2 {
				result = append(result, stemmed)
			} else {
				// Fall back to original token if stemming produces invalid result
				result = append(result, token)
			}
		}
	}

	return result
}

// isNumeric checks if a string consists only of digits
func isNumeric(s string) bool {
	for _, r := range s {
		if !unicode.IsDigit(r) {
			return false
		}
	}
	return len(s) > 0
}

// IsBinaryContent detects if content appears to be binary.
// It checks for null bytes and high ratio of non-printable characters.
func IsBinaryContent(content string) bool {
	if len(content) == 0 {
		return false
	}

	// Sample first 1024 bytes to check for binary content
	sampleSize := 1024
	if len(content) < sampleSize {
		sampleSize = len(content)
	}
	sample := content[:sampleSize]

	nonPrintable := 0
	for _, r := range sample {
		// Null byte is a strong indicator of binary
		if r == 0 {
			return true
		}
		// Count non-printable characters (excluding common whitespace)
		if !unicode.IsPrint(r) && r != '\n' && r != '\r' && r != '\t' {
			nonPrintable++
		}
	}

	// If more than 30% non-printable, consider it binary
	return float64(nonPrintable)/float64(len(sample)) > 0.30
}

// TokenizeJSON handles JSON-specific tokenization.
// It extracts keys and string values as tokens.
func TokenizeJSON(content string) []string {
	// First check if it's valid JSON
	var data interface{}
	if err := json.Unmarshal([]byte(content), &data); err != nil {
		// Not valid JSON, fall back to regular tokenization
		return TokenizeContent(content)
	}

	var tokens []string
	extractJSONTokens(data, &tokens)
	return FilterStopwords(tokens)
}

// extractJSONTokens recursively extracts tokens from JSON data structures.
func extractJSONTokens(data interface{}, tokens *[]string) {
	switch v := data.(type) {
	case map[string]interface{}:
		for key, value := range v {
			// Tokenize the key
			*tokens = append(*tokens, processIdentifier(key)...)
			// Recursively process value
			extractJSONTokens(value, tokens)
		}
	case []interface{}:
		for _, item := range v {
			extractJSONTokens(item, tokens)
		}
	case string:
		// Tokenize string values (without stopword filtering, done at end)
		*tokens = append(*tokens, tokenizeCode(v)...)
		// Numeric and boolean values are skipped
	}
}

// TokenizeQuery tokenizes a search query using the same pipeline as indexing.
// This ensures query tokens match indexed tokens for accurate BM25 scoring.
func TokenizeQuery(query string) []string {
	return TokenizeContent(query)
}
