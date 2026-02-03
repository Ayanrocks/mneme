package cli

import (
	"mneme/internal/config"
	"mneme/internal/constants"
	"mneme/internal/logger"
	"mneme/internal/storage"
	"mneme/internal/version"
	"os"
	"path"
	"runtime"

	"github.com/spf13/cobra"
)

var (
	verbose bool
	quiet   bool
)

// getLogLevelFromConfig attempts to load the config and return the log level.
// Returns empty string if config cannot be loaded, which will default to "info".
func getLogLevelFromConfig() string {
	cfg, err := config.LoadConfig()
	if err != nil {
		// Config not available yet (possibly not initialized), return empty to default to info
		return ""
	}
	return cfg.Logging.Level
}

var rootCmd = &cobra.Command{
	Use:   "mneme",
	Short: "Mneme - A powerful personal search engine",
	Long:  `Mneme - a powerful search engine to search for all the personal documents, repos, files`,

	PreRun: func(cmd *cobra.Command, args []string) {
		// Initialize logger after flags are parsed, with log level from config
		logLevel := getLogLevelFromConfig()
		logger.Init(verbose, quiet, false, logLevel)
	},

	Run: func(cmd *cobra.Command, args []string) {
		// User-facing output (clean, no timestamps)
		logger.Header("Welcome to Mneme! 🚀")
		logger.Print("Use 'mneme --help' to see available commands")
		logger.Blank()

		// Development logs (structured, with timestamps)
		logger.Debug("Initializing Mneme...")
		logger.Debug("Checking available commands...")
		logger.Debug("Application ready")
	},
}

var versionCmd = &cobra.Command{
	Use:     "version",
	Aliases: []string{"v"},
	Short:   "Show version information",
	Long:    `Display the current version of Mneme.`,

	Run: versionCmdExecute,
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initializing mneme setup with indexing",
	Long:  `Initializing mneme setup with indexing`,
	Run:   initCmdExecute,
}

func Execute() error {
	// Initialize logger with flags before executing
	// Use empty log level for initial startup, will be properly set in PreRun
	logger.Init(verbose, quiet, false, "")
	return rootCmd.Execute()
}

func init() {
	// Add logging flags
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose debug logging")
	rootCmd.PersistentFlags().BoolVarP(&quiet, "quiet", "q", false, "Enable quiet mode (only errors)")

	// Add commands to root
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(configCmd)
	rootCmd.AddCommand(indexCmd)
	rootCmd.AddCommand(findCmd)
}

// IsInitialized checks if the init command was run by verifying that the
// config file, the data directory, and all nested directories exist.
// Returns true if all paths exist, false if any is missing.
func IsInitialized() (bool, error) {
	// Check if the config file exists
	configExists, err := storage.FileExists(constants.ConfigPath)
	if err != nil {
		logger.Errorf("Error checking config file: %+v", err)
		return false, err
	}

	// Check if the data directory exists
	dirExists, err := storage.DirExists(constants.DirPath)
	if err != nil {
		logger.Errorf("Error checking data directory: %+v", err)
		return false, err
	}

	// Both must exist for init to be considered complete
	if !configExists || !dirExists {
		if !configExists {
			logger.Debug("Config file does not exist - init has not been run")
		}
		if !dirExists {
			logger.Debug("Data directory does not exist - init has not been run")
		}
		return false, nil
	}

	// Check if all nested directories exist
	nestedDirs := []string{"meta", "segments", "tombstones"}
	for _, dir := range nestedDirs {
		dirPath := path.Join(constants.DirPath, dir)
		exists, err := storage.DirExists(dirPath)
		if err != nil {
			logger.Errorf("Error checking nested directory %s: %+v", dir, err)
			return false, err
		}
		if !exists {
			logger.Debugf("Nested directory %s does not exist - init is incomplete", dir)
			return false, nil
		}
	}

	logger.Debug("Config file, data directory, and all nested directories exist - init was run")
	return true, nil
}

func versionCmdExecute(cmd *cobra.Command, args []string) {
	logger.Header("Mneme Version")
	logger.KeyValue("Version", version.MnemeVersion)
	logger.KeyValue("Storage Engine", version.MnemeStorageEngineVersion)
	logger.KeyValue("Go Version", runtime.Version()[2:])
	logger.Blank()
}

func initCmdExecute(cmd *cobra.Command, args []string) {
	// initialize the mneme cli with basic configuration setup.
	err := storage.InitMnemeStorage()
	if err != nil {
		logger.Error("Failed to initialize mneme storage")
		os.Exit(1)
		return
	}

	// initialize the mneme config
	err = storage.InitMnemeConfigStorage()
	if err != nil {
		logger.Error("Failed to initialize mneme config")
		os.Exit(1)
		return
	}

}
