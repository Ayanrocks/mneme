package core

type Segment struct {
	Docs          []Document         `json:"docs"`
	InvertedIndex map[string]Posting `json:"inverted_index"`
	TotalDocs     int                `json:"total_docs"`
	TotalTokens   int                `json:"total_tokens"`
	AvgDocLen     int                `json:"avg_doc_len"`
}
