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
Results are ranked by relevance using BM25 and Vector Space Model algorithms.

Use quotes to search for exact phrases:
  mneme find "aws region"       → matches the exact phrase "aws region"
  mneme find deploy production  → matches documents containing "deploy" or "production"
  mneme find "error handling" go → matches the phrase "error handling" and the word "go"`,
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
		logger.PrintError("Please provide a search query. Example: mneme find \"your query\"")
		return
	}

	// Build the raw query string from args
	queryString := strings.Join(args, " ")
	queryString = strings.TrimSpace(queryString)

	if queryString == "" {
		logger.PrintError("Search query cannot be empty. Example: mneme find \"your query\"")
		return
	}

	config, err := config.LoadConfig()
	if err != nil {
		logger.Errorf("Failed to load config: %+v", err)
		return
	}

	// Use args directly as search terms.
	// The shell already handles quote parsing:
	//   mneme find "aws region"     → args = ["aws region"]  (one phrase)
	//   mneme find deploy prod      → args = ["deploy", "prod"] (two words)
	//   mneme find "error handling" go → args = ["error handling", "go"]
	// Multi-word args (containing spaces) were quoted by the user = phrase search.
	// Single-word args are individual token matches.
	searchTerms := args

	// Build stemmed tokens for BM25/VSM by splitting all terms into individual words
	stemmedTokens := query.ParseQuery(queryString)

	if len(stemmedTokens) == 0 {
		logger.PrintError("No valid search tokens found in query: %s", queryString)
		return
	}

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
			logger.PrintError("No index found. Please run 'mneme index' to build the search index first.")
			return
		}

		pb.SetMessage("Ranking documents...")
		// Get ranked documents with scores
		rankedDocs = query.RankDocuments(segmentIndex, stemmedTokens, config.Search.DefaultLimit, &config.Ranking)
		pb.Complete()
	} else {
		// No progress bar - regular logging
		// Read segments index from the file system
		segmentIndex, err = storage.LoadSegmentIndex()
		if err != nil {
			logger.PrintError("No index found. Please run 'mneme index' to build the search index first.")
			return
		}

		// Get ranked documents with scores
		rankedDocs = query.RankDocuments(segmentIndex, stemmedTokens, config.Search.DefaultLimit, &config.Ranking)
	}

	if len(rankedDocs) == 0 {
		logger.PrintError("No documents found for query: %s", queryString)
		return
	}

	// Use searchTerms (which preserve quoted phrases) for snippet matching
	// This ensures "aws region" is matched as an exact phrase in document lines
	var results []*core.SearchResult
	for _, doc := range rankedDocs {
		result, err := display.FormatSearchResult(doc.Path, searchTerms, doc.Score)
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
