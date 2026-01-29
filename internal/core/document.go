package core

type Document struct {
	ID         uint   `json:"id"`
	Path       string `json:"path"`
	TokenCount uint   `json:"token_count"`
}

type Posting struct {
	DocID uint `json:"doc_id"`
	Freq  uint `json:"freq"`
}
