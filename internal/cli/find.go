package cli

import (
	"strings"

	"mneme/internal/config"
	"mneme/internal/core"
	"mneme/internal/display"
	"mneme/internal/logger"
	"mneme/internal/platform"
	"mneme/internal/query"
	"mneme/internal/storage"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var findCmd = &cobra.Command{
	Use:   "find",
	Short: "Find documents",
	Long: `Find documents matching the given query, showing relevant snippets.
Results are ranked by relevance using BM25 and Vector Space Model algorithms.
Fuzzy matching is automatically applied to catch typos and near-matches.

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

	cfg, err := config.LoadConfig()
	if err != nil {
		logger.Errorf("Failed to load config: %+v", err)
		return
	}

	// Load the index first to enable auto-correction
	var segmentIndex *core.Segment

	if display.ShouldShowProgress() {
		pb := display.NewProgressBar("Initializing", 0)
		pb.Start()
		pb.SetMessage("Loading index...")
		segmentIndex, err = storage.LoadSegmentIndex()
		pb.Complete()
	} else {
		segmentIndex, err = storage.LoadSegmentIndex()
	}

	if err != nil {
		logger.PrintError("No index found. Please run 'mneme index' to build the search index first.")
		return
	}

	// Build the query string from args directly
	queryString := strings.Join(args, " ")
	queryString = strings.TrimSpace(queryString)

	if queryString == "" {
		logger.PrintError("Search query cannot be empty.")
		return
	}

	// Parse the query string into stemmed tokens
	// Use the corrected query string if available
	stemmedTokens := query.ParseQuery(queryString)

	if len(stemmedTokens) == 0 {
		logger.PrintError("No valid search tokens found in query: %s", queryString)
		return
	}

	// Check if we should show progress bar for ranking
	var rankedDocs []core.RankedDocument

	if display.ShouldShowProgress() {
		pb := display.NewProgressBar("Searching", 0)
		pb.Start()
		pb.SetMessage("Ranking documents...")

		rankedDocs = query.RankDocuments(segmentIndex, stemmedTokens, cfg.Search.DefaultLimit, &cfg.Ranking)
		pb.Complete()
	} else {
		rankedDocs = query.RankDocuments(segmentIndex, stemmedTokens, cfg.Search.DefaultLimit, &cfg.Ranking)
	}

	if len(rankedDocs) == 0 {
		logger.PrintError("No documents found for query: %s", queryString)
		return
	}

	// Use original searchTerms (which preserve quoted phrases) for snippet matching?
	// Actually, we should use correctedArgs for snippet matching too, so we highlight "find" instead of "fnid".
	var results []*core.SearchResult
	for _, doc := range rankedDocs {
		// Use the terms that actually matched this document (including fuzzy expansions)
		// for precise highlighting.
		result, err := display.FormatSearchResult(doc.Path, doc.MatchedTerms, doc.Score)
		if err != nil {
			logger.Debugf("Failed to format result for %s: %v", doc.Path, err)
			continue
		}

		// Only include results that have actual text matches (snippets)
		// This filters out false positives from BM25 stemming
		if len(result.Snippets) > 0 {
			results = append(results, result)
		} else {
			// Fallback: if corrected terms didn't match (maybe stems vs raw mismatch),
			// use the actual terms that matched during ranking (including fuzzy expansions).
			result, err = display.FormatSearchResult(doc.Path, doc.MatchedTerms, doc.Score)
			if err == nil && len(result.Snippets) > 0 {
				results = append(results, result)
			}
		}
	}

	if len(results) == 0 {
		logger.PrintError("No matching documents found for: %s", queryString)
		return
	}

	// Print formatted results
	display.PrintResults(results, true, queryString)
}
