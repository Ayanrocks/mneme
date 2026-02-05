package core

// CrawlerOptions contains configuration for the file crawler
type CrawlerOptions struct {
	// IncludeExtensions is a list of file extensions to include (without the leading dot, e.g., "go", "py")
	// If empty, all extensions are included. If non-empty, only files with these extensions are included.
	IncludeExtensions []string

	// ExcludeExtensions is a list of file extensions to exclude (without the leading dot, e.g., "log", "tmp")
	// Files with these extensions will be skipped even if they match IncludeExtensions.
	ExcludeExtensions []string

	// SkipFolders is a list of folder names to skip entirely
	SkipFolders []string

	// MaxFilesPerFolder is the maximum number of files to process per folder
	// If a folder contains more files than this limit, it will be skipped
	// Set to 0 or negative to disable this limit
	MaxFilesPerFolder int

	// IncludeHidden determines whether to include hidden files/folders (those starting with .)
	IncludeHidden bool

	// SkipBinaryFiles determines whether to skip binary and non-readable file types
	// When true, files with extensions in BinaryExtensions will be skipped
	SkipBinaryFiles bool
}

// DefaultCrawlerOptions returns a CrawlerOptions with sensible defaults
func DefaultCrawlerOptions() CrawlerOptions {
	return CrawlerOptions{
		IncludeExtensions: []string{},
		ExcludeExtensions: []string{},
		SkipFolders:       []string{".git", "node_modules", ".svn", ".hg", "__pycache__", ".idea", ".vscode"},
		MaxFilesPerFolder: 0, // No limit by default
		IncludeHidden:     false,
	}
}
