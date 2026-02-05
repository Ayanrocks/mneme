package cli

import (
	"mneme/internal/config"
	"mneme/internal/constants"
	"mneme/internal/core"
	"mneme/internal/display"
	"mneme/internal/index"
	"mneme/internal/logger"
	"mneme/internal/storage"
	"mneme/internal/utils"

	"github.com/spf13/cobra"
)

var indexCmd = &cobra.Command{
	Use:   "index",
	Short: "Create an index",
	Long:  "Create an index using LSM-style batch processing to reduce memory usage",
	Run:   indexCmdExecute,
}

func indexCmdExecute(cmd *cobra.Command, args []string) {
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

	// Move existing segments to tombstones before re-indexing
	err = storage.MoveSegmentsToTombstones()
	if err != nil {
		logger.Errorf("Failed to move segments to tombstones: %+v", err)
		return
	}

	crawlerOptions := core.CrawlerOptions{
		IncludeExtensions: config.Sources.IncludeExtensions,
		ExcludeExtensions: config.Sources.ExcludeExtensions,
		SkipFolders:       config.Sources.Ignore,
		MaxFilesPerFolder: 0,
		IncludeHidden:     false,
		SkipBinaryFiles:   config.Index.SkipBinaryFiles,
	}

	// Use batch indexing to reduce memory usage
	batchConfig := core.DefaultBatchConfig()

	// Check if we should show progress bar (only when log level is "info")
	if display.ShouldShowProgress() {
		// Create progress bar - starts as indeterminate (spinner)
		pb := display.NewProgressBar("Indexing", 0)
		pb.Start()

		// Set up progress callback - this will be called by IndexBuilderBatched
		batchConfig.ProgressCallback = func(current, total int, message string) {
			pb.SetTotal(total)
			pb.SetCurrent(current)
			pb.SetMessage(message)
		}

		// Suppress regular logging during progress bar by setting NoLogging flag
		batchConfig.SuppressLogs = true

		manifest, err := index.IndexBuilderBatched(paths, &crawlerOptions, batchConfig)
		pb.Complete()

		if err != nil {
			logger.Errorf("Failed to build index: %+v", err)
			return
		}

		if manifest == nil {
			logger.Warn("No files were indexed")
			return
		}

		logger.Print("Indexing completed: %d chunks, %d docs, %d tokens",
			len(manifest.Chunks), manifest.TotalDocs, manifest.TotalTokens)
		CheckTombstonesAndHint()
	} else {
		// No progress bar - regular logging mode
		logger.Infof("Starting batch indexing (batch size: %d files)", batchConfig.BatchSize)

		manifest, err := index.IndexBuilderBatched(paths, &crawlerOptions, batchConfig)
		if err != nil {
			logger.Errorf("Failed to build index: %+v", err)
			return
		}

		if manifest == nil {
			logger.Warn("No files were indexed")
			return
		}

		logger.Infof("Indexing completed successfully: %d chunks, %d docs, %d tokens",
			len(manifest.Chunks), manifest.TotalDocs, manifest.TotalTokens)
		CheckTombstonesAndHint()
	}
}
