package query

import (
	"mneme/internal/core"
	"mneme/internal/index"
	"mneme/internal/logger"
)

func ParseQuery(queryString string) []string {
	logger.Info("Query string: " + queryString)

	// Use the same tokenization pipeline as indexing for BM25 consistency
	tokens := index.TokenizeQuery(queryString)

	return tokens
}

func FindQueryToken(segment *core.Segment, tokens []string) []string {
	logger.Info("Finding query token in segments ")

	docHits := make(map[uint]uint)

	for _, token := range tokens {
		postings, ok := segment.InvertedIndex[token]
		if !ok {
			continue
		}

		for _, posting := range postings {
			docHits[posting.DocID]++
		}
	}

	docPaths := []string{}

	logger.Debugf("Doc hits: %+v", docHits)

	for docId, hits := range docHits {
		for _, doc := range segment.Docs {
			if doc.ID == docId && hits > 0 {
				docPaths = append(docPaths, doc.Path)
			}
		}
	}

	return docPaths
}
