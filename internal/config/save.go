package config

import (
	"fmt"
	"mneme/internal/utils"
	"os"
	"path/filepath"

	"github.com/pelletier/go-toml/v2"
	"github.com/spf13/cobra"

	"mneme/internal/constants"
	"mneme/internal/core"
	"mneme/internal/logger"
)

func AddCmdExecute(cmd *cobra.Command, args []string) {
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

	// Validate the input path before expansion
	// Check for obviously invalid patterns
	if len(path) > 0 {
		hasLetterOrNumber := false
		hasPathSeparator := false
		hasTilde := false

		for _, c := range path {
			if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') {
				hasLetterOrNumber = true
			}
			if c == '/' || c == '\\' {
				hasPathSeparator = true
			}
			if c == '~' {
				hasTilde = true
			}
		}

		// If path has no letters/numbers AND no path separators AND no tilde, it's likely invalid
		// (e.g., "!!!", "###", "   ", etc.)
		if !hasLetterOrNumber && !hasPathSeparator && !hasTilde {
			logger.PrintError("Invalid path provided. Path should contain letters, numbers, or path separators.")
			logger.PrintRaw("Usage: mneme config add <path>")
			logger.PrintRaw("")
			logger.PrintRaw("Examples:")
			logger.PrintRaw("  mneme config add /path/to/directory")
			logger.PrintRaw("  mneme config add ~/Documents")
			logger.PrintRaw("  mneme config add ./relative/path")
			logger.PrintRaw("  mneme config add $HOME/Projects")
			return
		}
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

	// Validate that the path is not empty after expansion
	if expandedPath == "" || len(expandedPath) == 0 {
		logger.PrintError("Invalid path provided. Path cannot be empty.")
		logger.PrintRaw("Usage: mneme config add <path>")
		logger.PrintRaw("")
		logger.PrintRaw("Examples:")
		logger.PrintRaw("  mneme config add /path/to/directory")
		logger.PrintRaw("  mneme config add ~/Documents")
		logger.PrintRaw("  mneme config add ./relative/path")
		logger.PrintRaw("  mneme config add $HOME/Projects")
		return
	}

	// Validate that the path exists on the filesystem
	if _, err := os.Stat(expandedPath); os.IsNotExist(err) {
		logger.Errorf("Path does not exist: %s", expandedPath)
		logger.PrintRaw("Please provide a valid path that exists on the filesystem")
		logger.PrintRaw("")
		logger.PrintRaw("Examples:")
		logger.PrintRaw("  mneme config add /path/to/directory")
		logger.PrintRaw("  mneme config add ~/Documents")
		logger.PrintRaw("  mneme config add ./relative/path")
		logger.PrintRaw("  mneme config add $HOME/Projects")
		return
	} else if err != nil {
		logger.Errorf("Error checking path: %+v", err)
		return
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
		logger.PrintRaw("Please run 'mneme init' to initialize the configuration first")
		return
	}

	// Read the existing config file
	var config core.Config
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

func RemoveCmdExecute(cmd *cobra.Command, args []string) {
	logger.Info("removeCmdExecute")

	// Check if --all flag is set
	allFlag, err := cmd.Flags().GetBool("all")
	if err != nil {
		logger.Errorf("Failed to get --all flag: %+v", err)
		return
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
		logger.PrintRaw("Please run 'mneme init' to initialize the configuration first")
		return
	}

	// Read the existing config file
	var config core.Config
	configBytes, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Config file doesn't exist, nothing to remove
			logger.Warnf("Config file not found at: %s", configPath)
			logger.PrintError("No configuration found. Nothing to remove.")
			logger.PrintRaw("Please run 'mneme init' to initialize the configuration first")
			return
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

	// Handle --all flag
	if allFlag {
		// Check if there are any paths to remove
		if len(config.Sources.Paths) == 0 {
			logger.Info("No paths found in configuration. Nothing to remove.")
			return
		}

		// Display the paths that will be removed
		logger.Info("The following paths will be removed:")
		for _, path := range config.Sources.Paths {
			logger.PrintRaw("%s", "  - "+path)
		}
		logger.PrintRaw("")

		// First confirmation
		logger.PrintRaw("Are you sure you want to remove ALL paths? This action cannot be undone.")
		logger.PrintRaw("Type 'yes' to confirm:")

		var firstConfirmation string
		_, err := fmt.Scanln(&firstConfirmation)
		if err != nil {
			logger.Errorf("Failed to read confirmation: %+v", err)
			return
		}

		if firstConfirmation != "yes" {
			logger.Info("Operation cancelled.")
			return
		}

		// Second confirmation
		logger.PrintRaw("")
		logger.PrintRaw("This will permanently remove all indexed paths from your configuration.")
		logger.PrintRaw("Type 'yes' again to confirm:")

		var secondConfirmation string
		_, err = fmt.Scanln(&secondConfirmation)
		if err != nil {
			logger.Errorf("Failed to read confirmation: %+v", err)
			return
		}

		if secondConfirmation != "yes" {
			logger.Info("Operation cancelled.")
			return
		}

		// Remove all paths
		config.Sources.Paths = []string{}

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

		logger.Info("Successfully removed all paths from config.")
		return
	}

	// Original single path removal logic
	// Validate that a path was provided
	if len(args) == 0 {
		logger.Error("No path provided. Usage: mneme config remove <path>")
		logger.PrintRaw("")
		logger.PrintRaw("Examples:")
		logger.PrintRaw("  mneme config remove /path/to/directory")
		logger.PrintRaw("  mneme config remove ~/Documents")
		logger.PrintRaw("  mneme config remove ./relative/path")
		logger.PrintRaw("  mneme config remove $HOME/Projects")
		logger.PrintRaw("")
		logger.PrintRaw("Or use --all flag to remove all paths:")
		logger.PrintRaw("  mneme config remove --all")
		return
	}

	path := args[0]
	if path == "" {
		logger.Error("Path cannot be empty. Usage: mneme config remove <path>")
		logger.PrintRaw("")
		logger.PrintRaw("Examples:")
		logger.PrintRaw("  mneme config remove /path/to/directory")
		logger.PrintRaw("  mneme config remove ~/Documents")
		logger.PrintRaw("  mneme config remove ./relative/path")
		logger.PrintRaw("  mneme config remove $HOME/Projects")
		logger.PrintRaw("")
		logger.PrintRaw("Or use --all flag to remove all paths:")
		logger.PrintRaw("  mneme config remove --all")
		return
	}

	// Validate the input path before expansion
	// Check for obviously invalid patterns
	if len(path) > 0 {
		hasLetterOrNumber := false
		hasPathSeparator := false
		hasTilde := false

		for _, c := range path {
			if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') {
				hasLetterOrNumber = true
			}
			if c == '/' || c == '\\' {
				hasPathSeparator = true
			}
			if c == '~' {
				hasTilde = true
			}
		}

		// If path has no letters/numbers AND no path separators AND no tilde, it's likely invalid
		// (e.g., "!!!", "###", "   ", etc.)
		if !hasLetterOrNumber && !hasPathSeparator && !hasTilde {
			logger.PrintError("Invalid path provided. Path should contain letters, numbers, or path separators.")
			logger.PrintRaw("Usage: mneme config remove <path>")
			logger.PrintRaw("")
			logger.PrintRaw("Examples:")
			logger.PrintRaw("  mneme config remove /path/to/directory")
			logger.PrintRaw("  mneme config remove ~/Documents")
			logger.PrintRaw("  mneme config remove ./relative/path")
			logger.PrintRaw("  mneme config remove $HOME/Projects")
			logger.PrintRaw("")
			logger.PrintRaw("Or use --all flag to remove all paths:")
			logger.PrintRaw("  mneme config remove --all")
			return
		}
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

	// Validate that the path is not empty after expansion
	if expandedPath == "" || len(expandedPath) == 0 {
		logger.PrintError("Invalid path provided. Path cannot be empty.")
		logger.PrintRaw("Usage: mneme config remove <path>")
		logger.PrintRaw("")
		logger.PrintRaw("Examples:")
		logger.PrintRaw("  mneme config remove /path/to/directory")
		logger.PrintRaw("  mneme config remove ~/Documents")
		logger.PrintRaw("  mneme config remove ./relative/path")
		logger.PrintRaw("  mneme config remove $HOME/Projects")
		logger.PrintRaw("")
		logger.PrintRaw("Or use --all flag to remove all paths:")
		logger.PrintRaw("  mneme config remove --all")
		return
	}

	// Check if path exists in config
	pathFound := false
	for i, existingPath := range config.Sources.Paths {
		if existingPath == expandedPath {
			pathFound = true
			// Remove the path from the slice
			config.Sources.Paths = append(config.Sources.Paths[:i], config.Sources.Paths[i+1:]...)
			break
		}
	}

	if !pathFound {
		logger.Errorf("Path not found in config: %s", expandedPath)
		logger.PrintRaw("Please provide a path that exists in the configuration")
		logger.PrintRaw("")
		logger.PrintRaw("To see current configuration, run:")
		logger.PrintRaw("  mneme config show")
		logger.PrintRaw("")
		logger.PrintRaw("Examples:")
		logger.PrintRaw("  mneme config remove /path/to/directory")
		logger.PrintRaw("  mneme config remove ~/Documents")
		logger.PrintRaw("  mneme config remove ./relative/path")
		logger.PrintRaw("  mneme config remove $HOME/Projects")
		logger.PrintRaw("")
		logger.PrintRaw("Or use --all flag to remove all paths:")
		logger.PrintRaw("  mneme config remove --all")
		return
	}

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

	logger.Infof("Successfully removed path from config: %s", expandedPath)
}
