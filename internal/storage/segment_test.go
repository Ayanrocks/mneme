package storage

import (
	"encoding/json"
	"mneme/internal/core"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

// createTestSegment creates a segment with realistic test data
func createTestSegment(numDocs int, numTerms int) *core.Segment {
	docs := make([]core.Document, numDocs)
	for i := 0; i < numDocs; i++ {
		docs[i] = core.Document{
			ID:         uint(i),
			Path:       "/path/to/file" + string(rune('a'+i%26)) + ".go",
			TokenCount: uint(100 + i%500),
		}
	}

	invertedIndex := make(map[string][]core.Posting, numTerms)
	for i := 0; i < numTerms; i++ {
		term := "term" + string(rune('a'+i%26)) + string(rune('0'+i%10))
		postings := make([]core.Posting, numDocs/10+1)
		for j := range postings {
			postings[j] = core.Posting{
				DocID: uint(j * 10 % numDocs),
				Freq:  uint(1 + j%5),
			}
		}
		invertedIndex[term] = postings
	}

	return &core.Segment{
		Docs:          docs,
		InvertedIndex: invertedIndex,
		TotalDocs:     uint(numDocs),
		TotalTokens:   uint(numDocs * 150),
		AvgDocLen:     150,
	}
}

func TestSegmentProtobufConversion(t *testing.T) {
	original := createTestSegment(100, 50)

	// Convert to protobuf
	pbSegment := original.ToPB()
	require.NotNil(t, pbSegment)

	// Convert back
	restored := core.SegmentFromPB(pbSegment)
	require.NotNil(t, restored)

	// Verify all fields match
	assert.Equal(t, original.TotalDocs, restored.TotalDocs)
	assert.Equal(t, original.TotalTokens, restored.TotalTokens)
	assert.Equal(t, original.AvgDocLen, restored.AvgDocLen)
	assert.Equal(t, len(original.Docs), len(restored.Docs))
	assert.Equal(t, len(original.InvertedIndex), len(restored.InvertedIndex))

	// Check individual docs
	for i, doc := range original.Docs {
		assert.Equal(t, doc.ID, restored.Docs[i].ID)
		assert.Equal(t, doc.Path, restored.Docs[i].Path)
		assert.Equal(t, doc.TokenCount, restored.Docs[i].TokenCount)
	}

	// Check inverted index
	for term, postings := range original.InvertedIndex {
		restoredPostings, ok := restored.InvertedIndex[term]
		require.True(t, ok, "term %s not found", term)
		assert.Equal(t, len(postings), len(restoredPostings), "posting count mismatch for term %s", term)
	}
}

func TestSegmentProtobufMarshalUnmarshal(t *testing.T) {
	original := createTestSegment(100, 50)

	// Convert to protobuf and marshal
	pbSegment := original.ToPB()
	data, err := proto.Marshal(pbSegment)
	require.NoError(t, err)
	require.NotEmpty(t, data)

	// Unmarshal and convert back
	var restored core.Segment
	restoredPB := original.ToPB() // Create a fresh proto message
	err = proto.Unmarshal(data, restoredPB)
	require.NoError(t, err)

	restored = *core.SegmentFromPB(restoredPB)

	// Verify
	assert.Equal(t, original.TotalDocs, restored.TotalDocs)
	assert.Equal(t, len(original.Docs), len(restored.Docs))
}

func TestSegmentBinarySizeReduction(t *testing.T) {
	segment := createTestSegment(1000, 500)

	// Get binary size
	pbSegment := segment.ToPB()
	binaryData, err := proto.Marshal(pbSegment)
	require.NoError(t, err)

	// Get JSON size (rough estimate using stdlib)
	jsonData, err := json.MarshalIndent(segment, "", "  ")
	require.NoError(t, err)

	binarySize := len(binaryData)
	jsonSize := len(jsonData)

	t.Logf("Binary size: %d bytes", binarySize)
	t.Logf("JSON size: %d bytes", jsonSize)
	t.Logf("Reduction: %.1f%% smaller", float64(jsonSize-binarySize)/float64(jsonSize)*100)

	// Binary should be significantly smaller
	assert.Less(t, binarySize, jsonSize, "Binary format should be smaller than JSON")
	// Typically 50-75% smaller
	assert.Less(t, float64(binarySize), float64(jsonSize)*0.75, "Binary should be at least 25% smaller")
}

func BenchmarkSegmentToPB(b *testing.B) {
	segment := createTestSegment(1000, 500)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		segment.ToPB()
	}
}

func BenchmarkSegmentFromPB(b *testing.B) {
	segment := createTestSegment(1000, 500)
	pbSegment := segment.ToPB()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		core.SegmentFromPB(pbSegment)
	}
}

func BenchmarkMarshalJSON(b *testing.B) {
	segment := createTestSegment(1000, 500)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		json.MarshalIndent(segment, "", "  ")
	}
}

func BenchmarkMarshalProtobuf(b *testing.B) {
	segment := createTestSegment(1000, 500)
	pbSegment := segment.ToPB()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		proto.Marshal(pbSegment)
	}
}

func BenchmarkUnmarshalJSON(b *testing.B) {
	segment := createTestSegment(1000, 500)
	data, _ := json.MarshalIndent(segment, "", "  ")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var s core.Segment
		json.Unmarshal(data, &s)
	}
}

func BenchmarkUnmarshalProtobuf(b *testing.B) {
	segment := createTestSegment(1000, 500)
	pbSegment := segment.ToPB()
	data, _ := proto.Marshal(pbSegment)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var pb core.Segment
		pbMsg := segment.ToPB()
		proto.Unmarshal(data, pbMsg)
		pb = *core.SegmentFromPB(pbMsg)
		_ = pb
	}
}

func TestFileSaveLoadDirect(t *testing.T) {
	tempDir := t.TempDir()
	segmentDir := filepath.Join(tempDir, "segments")
	err := os.MkdirAll(segmentDir, 0755)
	require.NoError(t, err)

	original := createTestSegment(50, 25)

	// Save directly to temp path
	binaryPath := filepath.Join(segmentDir, "segment.idx")
	pbSegment := original.ToPB()
	data, err := proto.Marshal(pbSegment)
	require.NoError(t, err)

	err = os.WriteFile(binaryPath, data, 0644)
	require.NoError(t, err)

	// Load directly
	loadedData, err := os.ReadFile(binaryPath)
	require.NoError(t, err)

	var loadedPB = original.ToPB()
	err = proto.Unmarshal(loadedData, loadedPB)
	require.NoError(t, err)

	loaded := core.SegmentFromPB(loadedPB)

	assert.Equal(t, original.TotalDocs, loaded.TotalDocs)
	assert.Equal(t, len(original.Docs), len(loaded.Docs))
}
