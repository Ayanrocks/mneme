package config

import (
	"mneme/internal/utils"
	"os"
	"path/filepath"

	"github.com/pelletier/go-toml/v2"
	"github.com/spf13/cobra"

	"mneme/internal/constants"
	"mneme/internal/core"
	"mneme/internal/logger"
)

func addCmdExecute(cmd *cobra.Command, args []string) {
	logger.Info("addCmdExecute")

	// Validate that a path was provided
	if len(args) == 0 {
		logger.Error("No path provided. Usage: mneme config add <path>")
		return
	}

	path := args[0]
	if path == "" {
		logger.Error("Path cannot be empty. Usage: mneme config add <path>")
		return
	}

	// Expand the path (handle ~ and environment variables)
	expandedPath := os.ExpandEnv(path)

	// Convert to absolute path if it's relative
	if !filepath.IsAbs(expandedPath) {
		absPath, err := filepath.Abs(expandedPath)
		if err != nil {
			logger.Errorf("Failed to convert path to absolute: %+v", err)
			return
		}
		expandedPath = absPath
	}

	// Expand the config path
	configPath, err := utils.ExpandFilePath(constants.ConfigPath)
	if err != nil {
		logger.Errorf("Failed to expand config path: %+v", err)
		return
	}
	configDir := filepath.Dir(configPath)

	// Check if config directory exists
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		logger.Errorf("Config directory does not exist: %s", configDir)
		logger.Info("Please run 'mneme init' to initialize the configuration first")
		return
	}

	// Read the existing config file
	var config core.DefaultConfig
	configBytes, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Config file doesn't exist, use default config
			logger.Warnf("Config file not found at: %s", configPath)
			logger.Info("Using default configuration")
			config = DefaultConfig
		} else {
			logger.Errorf("Failed to read config file: %+v", err)
			return
		}
	} else {
		// Parse existing config
		if err := toml.Unmarshal(configBytes, &config); err != nil {
			logger.Errorf("Failed to parse config file: %+v", err)
			return
		}
	}

	// Check if path already exists in config
	for _, existingPath := range config.Sources.Paths {
		if existingPath == expandedPath {
			logger.Warnf("Path already exists in config: %s", expandedPath)
			return
		}
	}

	// Add the path to config
	config.Sources.Paths = append(config.Sources.Paths, expandedPath)

	// Marshal config back to TOML
	configBytes, err = toml.Marshal(config)
	if err != nil {
		logger.Errorf("Failed to marshal config to TOML: %+v", err)
		return
	}

	// Write config back to file
	if err := os.WriteFile(configPath, configBytes, 0644); err != nil {
		logger.Errorf("Failed to write config file: %+v", err)
		return
	}

	logger.Infof("Successfully added path to config: %s", expandedPath)
}
