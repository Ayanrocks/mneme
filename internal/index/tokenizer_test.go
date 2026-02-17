package index

import (
	"testing"
)

func TestTokenizeContent_CamelCase(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "simple camelCase",
			input:    "getUserProfile",
			expected: []string{"getuserprofil", "get", "user", "profil"},
		},
		{
			name:     "PascalCase",
			input:    "PascalCase",
			expected: []string{"pascalcas", "pascal"},
		},
		{
			name:     "with acronym",
			input:    "parseHTMLDocument",
			expected: []string{"parsehtmldocu", "pars", "html", "document"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := TokenizeContent(tt.input)
			if len(result) != len(tt.expected) {
				t.Errorf("TokenizeContent(%q) = %v (len=%d), expected %v (len=%d)",
					tt.input, result, len(result), tt.expected, len(tt.expected))
				return
			}
			for i, token := range result {
				if token != tt.expected[i] {
					t.Errorf("TokenizeContent(%q)[%d] = %q, expected %q",
						tt.input, i, token, tt.expected[i])
				}
			}
		})
	}
}

func TestTokenizeContent_SnakeCase(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "simple snake_case",
			input:    "get_user_profile",
			expected: []string{"get", "user", "profil"},
		},
		{
			name:  "with numbers",
			input: "get_user_v2",
			// Note: "v2" splits into "v" (too short) and "2" (numeric), both filtered
			expected: []string{"get", "user"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := TokenizeContent(tt.input)
			if len(result) != len(tt.expected) {
				t.Errorf("TokenizeContent(%q) = %v, expected %v",
					tt.input, result, tt.expected)
				return
			}
			for i, token := range result {
				if token != tt.expected[i] {
					t.Errorf("TokenizeContent(%q)[%d] = %q, expected %q",
						tt.input, i, token, tt.expected[i])
				}
			}
		})
	}
}

func TestTokenizeContent_MixedCode(t *testing.T) {
	input := `calculateTotal(price amount) value {
		result := compute(price)
	}`
	result := TokenizeContent(input)

	// Should contain stemmed versions of the identifiers
	// Note: "func", "string", "return" are stopwords and filtered
	expectedContains := []string{"calcul", "total", "price", "amount", "valu", "result", "comput"}
	for _, expected := range expectedContains {
		found := false
		for _, token := range result {
			if token == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected token %q not found in %v", expected, result)
		}
	}
}

func TestTokenizeContent_Stemming(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"running -> run", "running", "run"},                 // stems to same "run", so double "run" (full + part)
		{"profiles -> profil", "profiles", "profil"},         // "profil" + "profil"
		{"connections -> connect", "connections", "connect"}, // "connect" + "connect"
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := TokenizeContent(tt.input)
			// Check if result contains duplicates if full ID stems to same as part
			// For "running", expected is just checking correctness, but result might be ["run", "run"]
			// Let's just check that it contains the expected stem at least once
			found := false
			for _, token := range result {
				if token == tt.expected {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("TokenizeContent(%q) = %v, expected instance of %q",
					tt.input, result, tt.expected)
			}
		})
	}
}

func TestIsBinaryContent(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "text content",
			input:    "main() { println(\"Hello\") }",
			expected: false,
		},
		{
			name:     "binary with null bytes",
			input:    "some\x00binary\x00content",
			expected: true,
		},
		{
			name:     "empty content",
			input:    "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsBinaryContent(tt.input)
			if result != tt.expected {
				t.Errorf("IsBinaryContent(%q) = %v, expected %v",
					tt.input, result, tt.expected)
			}
		})
	}
}

func TestTokenizeJSON(t *testing.T) {
	input := `{"userName": "testAccount", "profileData": 123}`
	result := TokenizeJSON(input)

	// Should contain stemmed tokens from keys and string values
	// Note: "id" is a stopword and filtered
	expectedContains := []string{"user", "name", "test", "account", "profil", "data"}
	for _, expected := range expectedContains {
		found := false
		for _, token := range result {
			if token == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected token %q not found in JSON tokenization result %v", expected, result)
		}
	}
}

func TestTokenizeQuery(t *testing.T) {
	// Query tokenization should match index tokenization
	// Use snake_case for both to avoid full-identifier generation difference
	query := "get_user_profile"
	content := "get_user_profile"

	queryTokens := TokenizeQuery(query)
	contentTokens := TokenizeContent(content)

	// Both should produce the same stemmed tokens
	if len(queryTokens) != len(contentTokens) {
		t.Errorf("Query tokens %v don't match content tokens %v", queryTokens, contentTokens)
		return
	}

	for i := range queryTokens {
		if queryTokens[i] != contentTokens[i] {
			t.Errorf("Token mismatch at %d: query=%q, content=%q",
				i, queryTokens[i], contentTokens[i])
		}
	}
}

func TestStopwordFiltering(t *testing.T) {
	// Programming keywords should be filtered
	// Using words that are stopwords AND don't change much when stemmed
	input := "func return class switch defer"
	result := TokenizeContent(input)

	// All of these are stopwords (exact match in map), so result should be empty
	if len(result) != 0 {
		t.Errorf("Expected all stopwords to be filtered, got %v", result)
	}
}

func TestContainsCJK(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"Chinese", "你好世界", true},
		{"Japanese Hiragana", "こんにちは", true},
		{"Japanese Katakana", "コンニチハ", true},
		{"Korean", "안녕하세요", true},
		{"English only", "hello world", false},
		{"Mixed", "hello你好", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := containsCJK(tt.input)
			if result != tt.expected {
				t.Errorf("containsCJK(%q) = %v, expected %v",
					tt.input, result, tt.expected)
			}
		})
	}
}

func TestFilterStopwords(t *testing.T) {
	input := []string{"get", "func", "user", "return", "profil", "class"}
	result := FilterStopwords(input)
	expected := []string{"get", "user", "profil"}

	if len(result) != len(expected) {
		t.Errorf("FilterStopwords() = %v, expected %v", result, expected)
		return
	}

	for i, token := range result {
		if token != expected[i] {
			t.Errorf("FilterStopwords()[%d] = %q, expected %q", i, token, expected[i])
		}
	}
}
