# Changelog

All notable changes to this project will be documented in this file.

## [0.5.0] - 2026-02-19

### Added
- **Quoted Phrase Search**: Use quotes in `mneme find` to search for exact phrases (e.g., `mneme find "error handling" go`). Multi-word quoted arguments are treated as phrase matches in snippet highlighting, while individual words are matched separately.
- **Query Parsing Module**: New `internal/query/parse.go` with `ParseQueryInput` for structured tokenization of mixed phrase and keyword queries, with full test coverage.
- **Benchmark Suite**: New `internal/benchmark` package with a build-tagged (`//go:build benchmark`) test suite that measures crawl, index, search, read, and write throughput across corpus sizes of 100–5,000 files. Results are printed as a formatted table. Run via `make benchmarks`.
- **`TESTING.md`**: Added documentation for the benchmark targets.

### Changed
- **Find Command UX**: Improved `Long` description and `Example` fields for the `find` command with clear usage patterns for phrase vs. keyword search.
- **Search Result Filtering**: Results are now filtered to only include documents with actual text-matching snippets, reducing false positives from BM25 stemming.
- **Code Cleanup**: Removed unused exported functions, aligned comments, and applied CodeRabbit review suggestions across multiple packages.

### Fixed
- **Benchmark Isolation**: Fixed benchmark tests leaking into the user's real config and data directories by saving and restoring `constants.ConfigPath` alongside `constants.DirPath`.
- **Test Stability**: Fixed failing tests caused by stray quote characters in query argument handling.

---

## [0.4.0] - 2026-02-11

### Added
- **Scanner Buffer Constants**: New `ScannerInitialBufSize` (64 KB) and `ScannerMaxBufSize` (5 MB) constants in `internal/constants/fs.go` for configurable scanner limits.
- **Storage Tests**: Added tests in `internal/storage/fs_test.go` for new storage initialization paths.

### Changed
- **`MaxTokensPerDocument` Default**: Changed from a fixed value to `0`, meaning all tokens in a file are indexed by default. When set to a positive value, files exceeding the limit are skipped instead of truncated.

### Fixed
- **`bufio.Scanner: token too long`**: Resolved a crash when indexing files with very long lines (e.g., minified JSON) by increasing the scanner buffer from the default 64 KB up to 5 MB.
- **Log Level Not Evaluated at Runtime**: Fixed an issue where the configured log level was read at init time and never re-evaluated, causing log-level changes in `mneme.toml` to have no effect until restart.

---

## [0.3.0] - 2026-02-10

### Added
- **Windows Support**: Full cross-platform support for Windows, including platform-specific path handling and OS detection stored in the `VERSION` file for compatibility checks.
- **Platform Abstraction Layer**: New `internal/platform` package with platform-specific implementations (`platform_unix.go`, `platform_windows.go`) for portable path resolution and data directory management.
- **CI/CD Pipelines**: Added GitHub Actions workflows for automated testing (`test.yml`) and release builds (`release.yml`).
- **Build System**: Introduced a `Makefile` with standard build targets for streamlined local development and CI builds.
- **CLI Enhancements**: Added comprehensive descriptions and illustrative examples to all commands and subcommands for a better help experience.

### Changed
- **Configuration Restructure**: Moved configuration constants to `internal/constants` for clearer package responsibilities.
- **Crawler Hardening**: The crawler now skips Windows system files and directories (e.g., `$Recycle.Bin`, `System Volume Information`) using a case-insensitive lookup.
- **Storage Paths**: Updated `internal/storage/fs.go` to use the new platform abstraction layer for resolving data directories, replacing hardcoded Unix paths.
- **Documentation**: Updated `README.md` with data store details and improved build instructions.

### Fixed
- **Windows Temp Dir Fallback**: Resolved an issue where the Windows fallback path could cause permission errors by using the system temp directory instead.

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
