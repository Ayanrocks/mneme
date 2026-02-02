package core

import "time"

// Manifest tracks all segment chunks and their status.
// This is persisted as manifest.json in the segments directory.
type Manifest struct {
	Version     string      `json:"version"`
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
	TotalDocs   uint        `json:"total_docs"`
	TotalTokens uint        `json:"total_tokens"`
	AvgDocLen   uint        `json:"avg_doc_len"`
	Chunks      []ChunkInfo `json:"chunks"`
}

// ChunkInfo describes a single chunk file
type ChunkInfo struct {
	ID         int       `json:"id"`          // Chunk number (1, 2, 3...)
	Filename   string    `json:"filename"`    // e.g., "001.idx"
	Status     string    `json:"status"`      // "complete" or "in_progress"
	DocCount   uint      `json:"doc_count"`   // Number of documents in this chunk
	TokenCount uint      `json:"token_count"` // Number of unique tokens in this chunk
	CreatedAt  time.Time `json:"created_at"`
}

// ChunkStatus constants
const (
	ChunkStatusComplete   = "complete"
	ChunkStatusInProgress = "in_progress"
)

// ManifestVersion is the current version of the manifest format
const ManifestVersion = "1.0"

// NewManifest creates a new empty manifest
func NewManifest() *Manifest {
	now := time.Now()
	return &Manifest{
		Version:   ManifestVersion,
		CreatedAt: now,
		UpdatedAt: now,
		Chunks:    make([]ChunkInfo, 0),
	}
}

// AddChunk adds a new chunk to the manifest
func (m *Manifest) AddChunk(chunk ChunkInfo) {
	m.Chunks = append(m.Chunks, chunk)
	m.UpdatedAt = time.Now()
}

// GetCompleteChunks returns only chunks with "complete" status
func (m *Manifest) GetCompleteChunks() []ChunkInfo {
	complete := make([]ChunkInfo, 0)
	for _, chunk := range m.Chunks {
		if chunk.Status == ChunkStatusComplete {
			complete = append(complete, chunk)
		}
	}
	return complete
}

// MarkChunkComplete marks a chunk as complete by ID
func (m *Manifest) MarkChunkComplete(chunkID int) {
	for i := range m.Chunks {
		if m.Chunks[i].ID == chunkID {
			m.Chunks[i].Status = ChunkStatusComplete
			m.UpdatedAt = time.Now()
			return
		}
	}
}

// UpdateTotals recalculates total docs, tokens, and average doc length
func (m *Manifest) UpdateTotals() {
	var totalDocs, totalTokens uint
	for _, chunk := range m.Chunks {
		if chunk.Status == ChunkStatusComplete {
			totalDocs += chunk.DocCount
			totalTokens += chunk.TokenCount
		}
	}
	m.TotalDocs = totalDocs
	m.TotalTokens = totalTokens
	if totalDocs > 0 {
		m.AvgDocLen = totalTokens / totalDocs
	}
	m.UpdatedAt = time.Now()
}
