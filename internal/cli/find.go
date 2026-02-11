package cli

import (
	"mneme/internal/config"
	"mneme/internal/core"
	"mneme/internal/display"
	"mneme/internal/logger"
	"mneme/internal/platform"
	"mneme/internal/query"
	"mneme/internal/storage"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var findCmd = &cobra.Command{
	Use:   "find",
	Short: "Find documents",
	Long: `Find documents matching the given query, showing relevant snippets.
Results are ranked by relevance using BM25 and Vector Space Model algorithms.`,
	Example: `  mneme find "machine learning"
  mneme find python tutorial
  mneme find "error handling" in go`,
	Run: findCmdExecute,
}

func findCmdExecute(cmd *cobra.Command, args []string) {
	initialized, err := IsInitialized()
	if err != nil {
		logger.Errorf("Failed to check if initialized: %+v", err)
		return
	}

	if !initialized {
		logger.Error("Mneme is not initialized. Please run 'mneme init' first.")
		return
	}

	// Check platform compatibility and show hint if mismatch
	compatible, storedPlatform, err := storage.CheckPlatformCompatibility()
	if err == nil && !compatible {
		color.Yellow("⚠️  Index was created on '%s' but you're running on '%s'", storedPlatform, platform.Current())
		color.Cyan("   Paths in the index may not resolve correctly.")
		color.White("   Run 'mneme index' to re-index for this platform.\n")
	}

	if len(args) < 1 {
		logger.PrintError("Please provide a query to search for.")
		return
	}

	config, err := config.LoadConfig()
	if err != nil {
		logger.Errorf("Failed to load config: %+v", err)
		return
	}

	// get query from args
	queryString := strings.Join(args, " ")

	// Get stemmed tokens for BM25 ranking
	queryTokens := query.ParseQuery(queryString)

	// Check if we should show progress bar
	var segmentIndex *core.Segment
	var rankedDocs []core.RankedDocument

	if display.ShouldShowProgress() {
		pb := display.NewProgressBar("Searching", 0)
		pb.Start()
		pb.SetMessage("Loading index...")

		// Read segments index from the file system
		segmentIndex, err = storage.LoadSegmentIndex()
		if err != nil {
			pb.Complete()
			logger.Errorf("Failed to load segment index: %+v", err)
			return
		}

		pb.SetMessage("Ranking documents...")
		// Get ranked documents with scores
		rankedDocs = query.RankDocuments(segmentIndex, queryTokens, config.Search.DefaultLimit, &config.Ranking)
		pb.Complete()
	} else {
		// No progress bar - regular logging
		// Read segments index from the file system
		segmentIndex, err = storage.LoadSegmentIndex()
		if err != nil {
			logger.Errorf("Failed to load segment index: %+v", err)
			return
		}

		// Get ranked documents with scores
		rankedDocs = query.RankDocuments(segmentIndex, queryTokens, config.Search.DefaultLimit, &config.Ranking)
	}

	if len(rankedDocs) == 0 {
		logger.PrintError("No documents found for query: %s", queryString)
		return
	}

	// Extract original query words (not stemmed) for snippet matching
	originalQueryWords := strings.Fields(queryString)

	// Build search results with snippets - only include results with actual matches
	var results []*core.SearchResult
	for _, doc := range rankedDocs {
		result, err := display.FormatSearchResult(doc.Path, originalQueryWords, doc.Score)
		if err != nil {
			logger.Debugf("Failed to format result for %s: %v", doc.Path, err)
			continue
		}

		// Only include results that have actual text matches (snippets)
		// This filters out false positives from BM25 stemming
		if len(result.Snippets) > 0 {
			results = append(results, result)
		}
	}

	if len(results) == 0 {
		logger.PrintError("No matching documents found for: %s", queryString)
		return
	}

	// Print formatted results
	display.PrintResults(results, true, queryString)
}
