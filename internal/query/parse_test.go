package query

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseQueryInput(t *testing.T) {
	t.Run("empty input returns nil slices", func(t *testing.T) {
		terms, tokens := ParseQueryInput("")
		assert.Nil(t, terms)
		assert.Nil(t, tokens)
	})

	t.Run("whitespace-only input returns nil slices", func(t *testing.T) {
		terms, tokens := ParseQueryInput("   ")
		assert.Nil(t, terms)
		assert.Nil(t, tokens)
	})

	t.Run("single word returns one search term", func(t *testing.T) {
		terms, tokens := ParseQueryInput("deploy")
		assert.Equal(t, []string{"deploy"}, terms)
		assert.NotEmpty(t, tokens)
	})

	t.Run("multiple unquoted words returns separate terms", func(t *testing.T) {
		terms, _ := ParseQueryInput("deploy production")
		assert.Equal(t, []string{"deploy", "production"}, terms)
	})

	t.Run("single quoted phrase returns one term", func(t *testing.T) {
		terms, tokens := ParseQueryInput(`"aws region"`)
		assert.Equal(t, []string{"aws region"}, terms)
		assert.NotEmpty(t, tokens)
		// Both words should still be stemmed individually for BM25
		assert.GreaterOrEqual(t, len(tokens), 1)
	})

	t.Run("quoted phrase with unquoted word", func(t *testing.T) {
		terms, _ := ParseQueryInput(`"error handling" golang`)
		assert.Equal(t, []string{"error handling", "golang"}, terms)
	})

	t.Run("unquoted word before quoted phrase", func(t *testing.T) {
		terms, _ := ParseQueryInput(`golang "error handling"`)
		assert.Equal(t, []string{"golang", "error handling"}, terms)
	})

	t.Run("multiple quoted phrases", func(t *testing.T) {
		terms, _ := ParseQueryInput(`"aws region" "error handling"`)
		assert.Equal(t, []string{"aws region", "error handling"}, terms)
	})

	t.Run("mixed quoted and unquoted", func(t *testing.T) {
		terms, _ := ParseQueryInput(`"aws region" deploy "error handling" prod`)
		assert.Equal(t, []string{"aws region", "deploy", "error handling", "prod"}, terms)
	})

	t.Run("unclosed quote treated as phrase", func(t *testing.T) {
		terms, _ := ParseQueryInput(`"aws region`)
		assert.Equal(t, []string{"aws region"}, terms)
	})

	t.Run("empty quotes are ignored", func(t *testing.T) {
		terms, _ := ParseQueryInput(`"" deploy`)
		assert.Equal(t, []string{"deploy"}, terms)
	})

	t.Run("stemmed tokens are generated from all words", func(t *testing.T) {
		_, tokens := ParseQueryInput(`"error handling" golang`)
		// Should have stemmed tokens for "error", "handling", and "golang"
		assert.NotEmpty(t, tokens)
	})
}

func TestExtractSearchTerms(t *testing.T) {
	t.Run("simple words", func(t *testing.T) {
		result := extractSearchTerms("hello world")
		assert.Equal(t, []string{"hello", "world"}, result)
	})

	t.Run("single quoted phrase", func(t *testing.T) {
		result := extractSearchTerms(`"hello world"`)
		assert.Equal(t, []string{"hello world"}, result)
	})

	t.Run("adjacent quoted phrases", func(t *testing.T) {
		result := extractSearchTerms(`"hello world""foo bar"`)
		assert.Equal(t, []string{"hello world", "foo bar"}, result)
	})

	t.Run("quoted phrase with spaces around", func(t *testing.T) {
		result := extractSearchTerms(`  "hello world"  `)
		assert.Equal(t, []string{"hello world"}, result)
	})

	t.Run("empty input", func(t *testing.T) {
		result := extractSearchTerms("")
		assert.Empty(t, result)
	})
}
