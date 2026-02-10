package constants

const (
	DefaultBatchSize = 20000

	// ScannerInitialBufSize is the initial buffer size for bufio.Scanner (64KB)
	ScannerInitialBufSize = 64 * 1024
	// ScannerMaxBufSize is the maximum buffer size for bufio.Scanner (1MB).
	// Files with lines longer than this (e.g., minified JSON) will still fail.
	ScannerMaxBufSize = 1024 * 1024 * 5

	// TombstoneSizeThreshold is the size in bytes (100MB) above which users
	// are warned to run 'mneme clean' to reclaim disk space.
	TombstoneSizeThreshold = 100 * 1024 * 1024
)
