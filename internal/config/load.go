package config

import (
	"mneme/internal/logger"

	"github.com/spf13/cobra"
)

func showCmdExecute(cmd *cobra.Command, args []string) {
	logger.Info("showCmdExecute")

	// Load the config from the config path
}
