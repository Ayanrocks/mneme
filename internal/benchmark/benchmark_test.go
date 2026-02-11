//go:build benchmark

package benchmark

import (
	"fmt"
	"math/rand"
	"mneme/internal/constants"
	"mneme/internal/core"
	"mneme/internal/index"
	"mneme/internal/query"
	"mneme/internal/storage"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/fatih/color"
	"github.com/rodaine/table"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// BenchmarkResult holds metrics for a single corpus size
type BenchmarkResult struct {
	CorpusSize      int
	TotalDataSize   int64
	CrawlTime       time.Duration
	IndexTime       time.Duration
	SearchTime      time.Duration
	IndexThroughput float64 // files/sec
	InputThroughput float64 // MB/s (input processing speed)
	WriteSpeed      float64 // MB/s (disk write speed)
	ReadSpeed       float64 // MB/s (disk read speed)
	TotalTime       time.Duration
}

// random keywords to simulate code-like content
var keywords = []string{
	"func", "return", "if", "else", "var", "int", "string", "struct", "import", "package",
	"for", "range", "map", "chan", "go", "interface", "type", "const", "error", "nil",
	"true", "false", "break", "continue", "switch", "case", "default", "select", "defer",
}

// generateTestCorpus creates a set of files with random "code-like" content
func generateTestCorpus(t *testing.T, dir string, numFiles int) int64 {
	var totalSize int64

	// Ensure directory exists
	err := os.MkdirAll(dir, 0755)
	require.NoError(t, err)

	for i := 0; i < numFiles; i++ {
		filename := fmt.Sprintf("file_%04d.go", i)
		path := filepath.Join(dir, filename)

		// varied content length: 50 to 200 lines
		numLines := 50 + rand.Intn(150)
		var content string

		for j := 0; j < numLines; j++ {
			// Randomly indentation
			indent := ""
			if rand.Intn(2) == 0 {
				indent = "\t"
			}

			// Generate a line with 3-10 words
			numWords := 3 + rand.Intn(8)
			line := indent
			for k := 0; k < numWords; k++ {
				if rand.Intn(3) == 0 {
					// Use a keyword
					line += keywords[rand.Intn(len(keywords))] + " "
				} else {
					// Use a random identifier
					line += fmt.Sprintf("ident_%d ", rand.Intn(1000))
				}
			}
			content += line + "\n"
		}

		err := os.WriteFile(path, []byte(content), 0644)
		require.NoError(t, err)

		totalSize += int64(len(content))
	}

	return totalSize
}

func TestBenchmarkSuite(t *testing.T) {
	// Corpus sizes to test
	sizes := []int{100, 500, 1000, 5000}
	results := make([]BenchmarkResult, 0, len(sizes))

	// Save original DirPath to restore after tests
	originalDirPath := constants.DirPath
	defer func() {
		constants.DirPath = originalDirPath
	}()

	fmt.Println("\nRunning Mneme Benchmarks...")
	fmt.Println("============================")

	for _, size := range sizes {
		fmt.Printf("\nBenchmarking corpus size: %d files...\n", size)

		// Create a fresh temp directory for each run
		tempDir := t.TempDir()
		corpusDir := filepath.Join(tempDir, "corpus")
		dataDir := filepath.Join(tempDir, "data")

		// Override constants.DirPath to use our temp data directory
		constants.DirPath = dataDir

		// Initialize storage (creates segments dir etc)
		err := storage.InitMnemeStorage()
		require.NoError(t, err)

		// 1. Generate Corpus
		start := time.Now()
		totalBytes := generateTestCorpus(t, corpusDir, size)
		genTime := time.Since(start)
		fmt.Printf("Generated corpus in %v (%d bytes)\n", genTime, totalBytes)

		res := BenchmarkResult{
			CorpusSize:    size,
			TotalDataSize: totalBytes,
		}

		// 2. Measure Crawl
		crawlerOpts := core.DefaultCrawlerOptions()
		start = time.Now()
		paths, err := storage.Crawler(corpusDir, crawlerOpts)
		res.CrawlTime = time.Since(start)
		require.NoError(t, err)
		assert.Equal(t, size, len(paths))

		// 3. Measure Indexing (including I/O)
		batchConfig := core.DefaultBatchConfig()
		batchConfig.SuppressLogs = true // Reduce noise
		// Ensure max tokens allows our files
		batchConfig.IndexConfig.MaxTokensPerDocument = 100000

		start = time.Now()
		manifest, err := index.IndexBuilderBatched(paths, &crawlerOpts, batchConfig)
		res.IndexTime = time.Since(start)
		require.NoError(t, err)
		require.NotNil(t, manifest)

		// Calculate indexing throughput
		res.IndexThroughput = float64(size) / res.IndexTime.Seconds()
		res.InputThroughput = float64(totalBytes) / 1024 / 1024 / res.IndexTime.Seconds()

		// 4. Measure Load (Read Speed)
		start = time.Now()
		segment, err := storage.LoadAllChunks()
		loadTime := time.Since(start)
		require.NoError(t, err)
		require.NotNil(t, segment)

		// Approximate Read Speed (MB/s) - using total index size (approx)
		// We can get actual size from checking file sizes in segments dir
		var indexSize int64
		filepath.Walk(filepath.Join(dataDir, "segments"), func(_ string, info os.FileInfo, _ error) error {
			if info != nil && !info.IsDir() {
				indexSize += info.Size()
			}
			return nil
		})
		res.ReadSpeed = float64(indexSize) / 1024 / 1024 / loadTime.Seconds()

		// Approximate Write Speed (MB/s) - using index size / index time (conservative)
		// Index time includes tokenization, so this is "Indexing Throughput (MB/s)"
		res.WriteSpeed = float64(indexSize) / 1024 / 1024 / res.IndexTime.Seconds()

		// 5. Measure Search
		// Perform multiple searches to get average
		rankingCfg := &core.RankingConfig{
			BM25Weight: 0.7,
			VSMWeight:  0.3,
		}

		start = time.Now()
		numSearches := 100
		for i := 0; i < numSearches; i++ {
			// Search for common keywords
			term := keywords[rand.Intn(len(keywords))]
			query.RankDocuments(segment, []string{term}, 10, rankingCfg)
		}
		res.SearchTime = time.Since(start) / time.Duration(numSearches)

		res.TotalTime = res.CrawlTime + res.IndexTime + res.SearchTime // + genTime excluded
		results = append(results, res)
	}

	// Print Results Table
	printResultsTable(results)
}

func printResultsTable(results []BenchmarkResult) {
	headerFmt := color.New(color.FgGreen, color.Underline).SprintfFunc()
	columnFmt := color.New(color.FgYellow).SprintfFunc()

	tbl := table.New("Files", "Data Size", "Crawl", "Index", "Search (avg)", "Files/sec", "Input (MB/s)", "Write (MB/s)", "Read (MB/s)")
	tbl.WithHeaderFormatter(headerFmt).WithFirstColumnFormatter(columnFmt)

	for _, r := range results {
		tbl.AddRow(
			r.CorpusSize,
			formatBytes(r.TotalDataSize),
			r.CrawlTime.Round(time.Millisecond),
			r.IndexTime.Round(time.Millisecond),
			r.SearchTime.Round(time.Microsecond),
			fmt.Sprintf("%.0f", r.IndexThroughput),
			fmt.Sprintf("%.2f", r.InputThroughput),
			fmt.Sprintf("%.2f", r.WriteSpeed),
			fmt.Sprintf("%.2f", r.ReadSpeed),
		)
	}

	fmt.Println("\nBenchmark Results:")
	tbl.Print()
}

func formatBytes(b int64) string {
	const unit = 1000
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "kMGTPE"[exp])
}
