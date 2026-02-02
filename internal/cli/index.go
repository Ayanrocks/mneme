package cli

import (
	"mneme/internal/config"
	"mneme/internal/constants"
	"mneme/internal/index"
	"mneme/internal/logger"
	"mneme/internal/storage"
	"mneme/internal/utils"

	"github.com/spf13/cobra"
)

var indexCmd = &cobra.Command{
	Use:   "index",
	Short: "Create an index",
	Long:  "Create an index",
	Run:   indexCmdExecute,
}

func indexCmdExecute(cmd *cobra.Command, args []string) {
	logger.Info("indexCmdExecute")

	initialized, err := IsInitialized()
	if err != nil {
		logger.Errorf("Failed to check if initialized: %+v", err)
		return
	}

	if !initialized {
		logger.Error("Mneme is not initialized. Please run 'mneme init' first.")
		return
	}

	config, err := config.LoadConfig()
	if err != nil {
		logger.Errorf("Failed to load config: %+v", err)
		return
	}

	// after the config is loaded, check and get the paths
	paths := config.Sources.Paths
	if len(paths) == 0 {
		logger.Error("No paths found in config")
		return
	}

	// Expand the data directory path (handle ~ expansion)
	dataDir, err := utils.ExpandFilePath(constants.DirPath)
	if err != nil {
		logger.Errorf("Failed to expand data directory path: %+v", err)
		return
	}

	// check if an existing lock is stale before trying to acquire
	if err := storage.CheckLock(dataDir); err != nil {
		// A lock exists, check if it's stale
		isStale, staleErr := storage.IsLockStale(dataDir)
		if staleErr != nil {
			logger.Errorf("Failed to check if lock is stale: %+v", staleErr)
			return
		}

		if isStale {
			logger.Warn("Found stale lock, clearing it...")
			if releaseErr := storage.ReleaseLock(dataDir); releaseErr != nil {
				logger.Errorf("Failed to release stale lock: %+v", releaseErr)
				return
			}
		} else {
			// Lock is held by an active process
			logger.Errorf("Failed to acquire lock: %+v", err)
			return
		}
	}

	// Now acquire the lock
	err = storage.AcquireLock(dataDir)
	if err != nil {
		logger.Errorf("Failed to acquire lock: %+v", err)
		return
	}

	// defer the release of the lock
	defer storage.ReleaseLock(dataDir)

	// TODO: check if the segments exists, then clear the directory and start fresh segmenting.
	// TO be done later

	crawlerOptions := storage.CrawlerOptions{
		IncludeExtensions: config.Sources.IncludeExtensions,
		SkipFolders:       config.Sources.Ignore,
		MaxFilesPerFolder: 0,
		IncludeHidden:     false,
	}

	// Start reading the files from the paths
	segmentIndex := index.IndexBuilder(paths, &crawlerOptions)

	// save the segment index to the data directory
	err = storage.SaveSegmentIndex(segmentIndex)
	if err != nil {
		logger.Errorf("Failed to save segment index: %+v", err)
		return
	}

	logger.Info("Indexing completed successfully")
}
