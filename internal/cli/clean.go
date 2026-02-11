package cli

import (
	"mneme/internal/constants"
	"mneme/internal/logger"
	"mneme/internal/storage"

	"github.com/spf13/cobra"
)

var cleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "Clean up tombstones folder to free disk space",
	Long: `Permanently delete old index files from the tombstones folder.
	
When you run 'mneme index', old index files are moved to the tombstones folder
instead of being deleted. This prevents accidental data loss. Run 'mneme clean'
to permanently remove these files and free up disk space.`,
	Example: `  mneme clean`,
	Run:     cleanCmdExecute,
}

func cleanCmdExecute(cmd *cobra.Command, args []string) {
	initialized, err := IsInitialized()
	if err != nil {
		logger.Errorf("Failed to check if initialized: %+v", err)
		return
	}

	if !initialized {
		logger.Error("Mneme is not initialized. Please run 'mneme init' first.")
		return
	}

	// Check current tombstones size
	currentSize, err := storage.GetTombstonesSize()
	if err != nil {
		logger.Errorf("Failed to get tombstones size: %+v", err)
		return
	}

	if currentSize == 0 {
		logger.Print("Tombstones folder is already empty. Nothing to clean.")
		return
	}

	logger.Print("Cleaning tombstones folder...")
	logger.Print("Current size: %s", storage.FormatBytes(currentSize))

	// Clear tombstones
	freedBytes, deletedCount, err := storage.ClearTombstones()
	if err != nil {
		logger.Errorf("Failed to clear tombstones: %+v", err)
		return
	}

	logger.Print("✓ Cleaned %d files, freed %s", deletedCount, storage.FormatBytes(freedBytes))
}

// CheckTombstonesAndHint checks if tombstones folder is large and prints a hint
func CheckTombstonesAndHint() {
	size, err := storage.GetTombstonesSize()
	if err != nil {
		logger.Debugf("Failed to check tombstones size: %+v", err)
		return
	}

	if size > constants.TombstoneSizeThreshold {
		logger.Print("")
		logger.Print("💡 Tombstones folder is using %s. Run 'mneme clean' to free up space.", storage.FormatBytes(size))
	}
}
