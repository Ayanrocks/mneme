package core

import (
	"mneme/internal/core/pb"
)

type Segment struct {
	Docs          []Document           `json:"docs"`
	InvertedIndex map[string][]Posting `json:"inverted_index"`
	TotalDocs     uint                 `json:"total_docs"`
	TotalTokens   uint                 `json:"total_tokens"`
	AvgDocLen     uint                 `json:"avg_doc_len"`
}

// ToPB converts a Segment to its protobuf representation
func (s *Segment) ToPB() *pb.Segment {
	pbDocs := make([]*pb.Document, len(s.Docs))
	for i, doc := range s.Docs {
		pbDocs[i] = &pb.Document{
			Id:         uint32(doc.ID),
			Path:       doc.Path,
			TokenCount: uint32(doc.TokenCount),
		}
	}

	pbIndex := make(map[string]*pb.PostingList, len(s.InvertedIndex))
	for term, postings := range s.InvertedIndex {
		pbPostings := make([]*pb.Posting, len(postings))
		for i, p := range postings {
			pbPostings[i] = &pb.Posting{
				DocId: uint32(p.DocID),
				Freq:  uint32(p.Freq),
			}
		}
		pbIndex[term] = &pb.PostingList{Postings: pbPostings}
	}

	return &pb.Segment{
		Docs:          pbDocs,
		InvertedIndex: pbIndex,
		TotalDocs:     uint32(s.TotalDocs),
		TotalTokens:   uint32(s.TotalTokens),
		AvgDocLen:     uint32(s.AvgDocLen),
	}
}

// FromPB creates a Segment from its protobuf representation
func SegmentFromPB(pbSeg *pb.Segment) *Segment {
	docs := make([]Document, len(pbSeg.Docs))
	for i, pbDoc := range pbSeg.Docs {
		docs[i] = Document{
			ID:         uint(pbDoc.Id),
			Path:       pbDoc.Path,
			TokenCount: uint(pbDoc.TokenCount),
		}
	}

	invertedIndex := make(map[string][]Posting, len(pbSeg.InvertedIndex))
	for term, pbList := range pbSeg.InvertedIndex {
		postings := make([]Posting, len(pbList.Postings))
		for i, pbP := range pbList.Postings {
			postings[i] = Posting{
				DocID: uint(pbP.DocId),
				Freq:  uint(pbP.Freq),
			}
		}
		invertedIndex[term] = postings
	}

	return &Segment{
		Docs:          docs,
		InvertedIndex: invertedIndex,
		TotalDocs:     uint(pbSeg.TotalDocs),
		TotalTokens:   uint(pbSeg.TotalTokens),
		AvgDocLen:     uint(pbSeg.AvgDocLen),
	}
}
