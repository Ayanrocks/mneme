package config

import (
	"os"

	"github.com/spf13/cobra"

	"mneme/internal/constants"
	"mneme/internal/logger"
	"mneme/internal/utils"
)

func showCmdExecute(cmd *cobra.Command, args []string) {
	logger.Info("showCmdExecute")

	// Expand the config path (handle ~)
	configPath, err := utils.ExpandFilePath(constants.ConfigPath)
	if err != nil {
		logger.Errorf("Failed to expand config path: %+v", err)
		return
	}

	// Read the config file
	configBytes, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			logger.Warnf("Config file not found at: %s", configPath)
			logger.Info("Using default configuration")
			configStr, err := DefaultConfigWriter()
			if err != nil {
				logger.Errorf("Failed to write default config: %+v", err)
				return
			}
			configBytes = []byte(configStr)
		} else {
			logger.Errorf("Failed to read config file: %+v", err)
			return
		}
	}

	// Pretty print the config
	if err := utils.PrettyPrintConfig(configBytes); err != nil {
		logger.Errorf("Failed to pretty print config: %+v", err)
	}
}
