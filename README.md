
## Code Folder Structure

```aiignore

mneme/
├── cmd/
│   └── mneme/
│       └── main.go            # Entry point (wires CLI, no logic)
│
├── internal/
│   ├── cli/                   # CLI commands & argument parsing
│   │   ├── root.go            # Root command definition
│   │   ├── init.go            # `mneme init`
│   │   ├── find.go            # `mneme find`
│   │   ├── index.go           # `mneme index`
│   │   └── status.go          # `mneme status`
│   │
│   ├── core/                  # Domain types (pure, no IO)
│   │   ├── document.go
│   │   ├── metadata.go
│   │   └── types.go
│   │
│   ├── ingest/                # Data ingestion & normalisation
│   │   ├── fs.go              # Filesystem ingestion
│   │   ├── git.go             # Git repository ingestion
│   │   ├── stdin.go           # STDIN ingestion
│   │   ├── pdf.go             # PDF text extraction
│   │   └── normalise.go       # Text cleanup & token prep
│   │
│   ├── index/                 # Inverted index implementation
│   │   ├── builder.go         # Builds in-memory indexes
│   │   ├── tokenizer.go       # Tokenisation logic
│   │   ├── postings.go        # Token → docID mappings
│   │   ├── segment.go         # Immutable index segments
│   │   └── merge.go           # Segment compaction
│   │
│   ├── query/                 # Query parsing & execution
│   │   ├── parse.go
│   │   ├── plan.go
│   │   └── execute.go
│   │
│   ├── rank/                  # Scoring & ranking heuristics
│   │   ├── tfidf.go
│   │   ├── recency.go
│   │   └── score.go
│   │
│   ├── storage/               # Filesystem paths & persistence
│   │   ├── paths.go           # XDG path resolution
│   │   ├── init.go            # Directory creation & checks
│   │   ├── segments.go        # Segment persistence
│   │   ├── metadata.go        # Metadata storage
│   │   ├── tombstone.go       # Deletions tracking
│   │   └── lock.go            # Writer locks
│   │
│   ├── watcher/               # Filesystem change detection
│   │   └── watcher.go
│   │
│   ├── server/                # Local HTTP server (optional)
│   │   ├── http.go
│   │   ├── handlers.go
│   │   └── middleware.go
│   │
│   └── util/                  # Shared helpers (small, boring)
│
├── data/                      # Test fixtures & sample documents
│
├── scripts/                   # Dev & build helpers
│
├── docs/                      # Design & architecture notes
│
├── go.mod
├── go.sum
└── README.md


```

## Mneme System Folder Structure

```aiignore
mneme/
├── meta/
│   ├── documents.db        # Document ID ↔ metadata mappings
│   └── instance_id         # Unique identifier for this Mneme instance
│
├── segments/
│   ├── segment_0001.idx    # Immutable inverted index segment
│   ├── segment_0002.idx
│   └── ...
│
├── tombstones/
│   └── deleted.ids         # Records of deleted or superseded documents
│
├── lock/
│   └── mneme.lock          # Prevents concurrent writers
│
└── VERSION                 # Storage format version

```