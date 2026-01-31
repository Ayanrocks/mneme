package cli

import (
	"mneme/internal/logger"
	"mneme/internal/query"
	"mneme/internal/storage"
	"strings"

	"github.com/spf13/cobra"
)

var findCmd = &cobra.Command{
	Use:   "find",
	Short: "Find documents",
	Long:  "Find documents",
	Run:   findCmdExecute,
}

func findCmdExecute(cmd *cobra.Command, args []string) {
	logger.Info("findCmdExecute")

	initialized, err := IsInitialized()
	if err != nil {
		logger.Errorf("Failed to check if initialized: %+v", err)
		return
	}

	if !initialized {
		logger.Error("Mneme is not initialized. Please run 'mneme init' first.")
		return
	}
	logger.Print("args: %+v", args)

	if len(args) < 1 {
		logger.PrintError("Please provide a query to search for.")
		return
	}

	// get query from args
	queryString := strings.Join(args, " ")

	logger.Info("Query string: " + queryString)

	queryTokens := query.ParseQuery(queryString)

	// Read segments index from the file system
	segmentIndex, err := storage.LoadSegmentIndex()
	if err != nil {
		logger.Errorf("Failed to load segment index: %+v", err)
		return
	}

	foundedDocPaths := query.FindQueryToken(segmentIndex, queryTokens)

	if len(foundedDocPaths) == 0 {
		logger.Info("No documents found for query: " + queryString)
		return
	}

	for _, docPath := range foundedDocPaths {
		logger.Info("Found doc path: " + docPath)
	}
}
