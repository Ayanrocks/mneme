package cli

import (
	"mneme/internal/logger"

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

	
}
