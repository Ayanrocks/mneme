package core

// RankedDocument represents a search result with its score
type RankedDocument struct {
	DocID uint
	Path  string
	Score float64
}

// GetScore implements utils.Scored interface
func (r RankedDocument) GetScore() float64 {
	return r.Score
}

// TFIDFVector represents a document or query as a TF-IDF weighted vector
type TFIDFVector struct {
	Weights map[string]float64
	Norm    float64 // L2 norm for cosine similarity
}

// SearchResult represents a formatted search result with snippet
type SearchResult struct {
	DocPath    string
	Score      float64
	Snippets   []Snippet
	MatchCount int
}

// Snippet represents a preview of the matched content
type Snippet struct {
	LineNumber int
	Content    string
	Highlights []HighlightRange
}

// HighlightRange marks the start and end positions of a match within a snippet
type HighlightRange struct {
	Start int
	End   int
}
