package core

type Segment struct {
	Docs          []Document           `json:"docs"`
	InvertedIndex map[string][]Posting `json:"inverted_index"`
	TotalDocs     uint                 `json:"total_docs"`
	TotalTokens   uint                 `json:"total_tokens"`
	AvgDocLen     uint                 `json:"avg_doc_len"`
}
