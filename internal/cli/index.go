package cli

import (
	"mneme/internal/config"
	"mneme/internal/constants"
	"mneme/internal/logger"
	"mneme/internal/storage"

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

	// check if a lock has been acquired in the data directory
	err = storage.AcquireLock(constants.DirPath)
	if err != nil {
		logger.Errorf("Failed to acquire lock: %+v", err)
		return
	}

	// defer the release of the lock
	defer storage.ReleaseLock(constants.DirPath)

	// check if the lock is stale
	isStale, err := storage.IsLockStale(constants.DirPath)
	if err != nil {
		logger.Errorf("Failed to check if lock is stale: %+v", err)
		return
	}

	if isStale {
		logger.Warn("Lock is stale. Please run 'mneme unlock' to remove it.")
		return
	}

	// TODO: check if the segments exists, then clear the directory and start fresh segmenting.
	// TO be done later

	// Start reading the files from the paths
	
	

}
