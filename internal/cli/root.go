package cli

import (
	"mneme/internal/logger"
	"mneme/internal/storage"
	"mneme/internal/version"
	"os"
	"runtime"

	"github.com/spf13/cobra"
)

var (
	verbose bool
	quiet   bool
)

var rootCmd = &cobra.Command{
	Use:   "mneme",
	Short: "Mneme - A powerful personal search engine",
	Long:  `Mneme - a powerful search engine to search for all the personal documents, repos, files`,

	PreRun: func(cmd *cobra.Command, args []string) {
		// Initialize logger after flags are parsed
		logger.Init(verbose, quiet, false)
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
	logger.Init(verbose, quiet, false)
	return rootCmd.Execute()
}

func init() {
	// Add logging flags
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose debug logging")
	rootCmd.PersistentFlags().BoolVarP(&quiet, "quiet", "q", false, "Enable quiet mode (only errors)")

	// Add commands to root
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(ConfigCmd)
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
