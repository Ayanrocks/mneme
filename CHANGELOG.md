# Changelog

All notable changes to this project will be documented in this file.

## [0.2.0] - 2026-02-08

### Added
- **Pluggable Ingestor System**: Introduced `Ingestor` interface to support multiple file sources, enabling future extensions beyond the local filesystem.
- **Batch Indexing**: Implemented LSM-tree inspired batch processing for index creation. This significantly reduces memory usage during indexing by writing numbered chunks and maintaining a `manifest.json`.
- **BM25 & VSM Ranking**: Integrated BM25 and Vector Space Model algorithms to improve search result relevance and ranking.
- **Binary Segment Storage**: Transititioned segment storage to Protocol Buffers for improved performance and reduced disk footprint.
- **Tombstones & Cleanup**: Added `mneme clean` command and a tombstone mechanism to safely manage and remove obsolete index files.
- **Multilingual Tokenization**: Enhanced tokenization logic to better support multi-lingual text analysis.
- **Configurable Logging**: Added log level configuration (defaulting to `info`) to reduce noise.
- **Binary File Skipping**: The crawler now automatically identifies and skips binary files (e.g., images, videos) to prevent indexing garbage data.
- **Exclude Extensions**: new configuration option `exclude_extensions` to explicitly ignore specific file types.

### Changed
- **Core Refactor**: Moved primary data structures to `internal/core` to improve modularity and reduce circular dependencies.
- **Progress Bar**: Unified the progress bar implementation to be "sticky" at the bottom of the terminal, providing a cleaner UX.
- **Search Highlighting**: Improved snippet generation to accurately highlight partial matches in search results.
- **Configuration**: Cleanup of unused ranking weights in the configuration file.

### Fixed
- **Module Dependencies**: Resolved indirect dependency issues in `go.mod`.
- **Test Stability**: Fixed flaky tests in `config` and `logger` packages related to environment permissions and default values.
