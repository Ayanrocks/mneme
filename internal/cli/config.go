package cli

import (
	"mneme/internal/config"

	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Configuration commands",
	Long:  `View and modify Mneme configuration, including managing indexed paths.`,
	Example: `  mneme config show
  mneme config add ~/Documents
  mneme config remove ~/Documents`,
}

var showCmd = &cobra.Command{
	Use:     "show",
	Short:   "Show configuration values",
	Long:    `Display the current configuration values, including indexed paths and settings.`,
	Example: `  mneme config show`,
	Run:     config.ShowCmdExecute,
}

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Add path to index",
	Long:  `Add a new directory path to be indexed by Mneme.`,
	Example: `  mneme config add ~/Documents
  mneme config add /path/to/repo`,
	Run: config.AddCmdExecute,
}

var removeCmd = &cobra.Command{
	Use:   "remove",
	Short: "Remove path from indexing",
	Long:  `Remove a directory path from Mneme's indexing list.`,
	Example: `  mneme config remove ~/Documents
  mneme config remove --all`,
	Run: config.RemoveCmdExecute,
}

func init() {
	removeCmd.Flags().BoolP("all", "a", false, "Remove all paths")

	configCmd.AddCommand(showCmd)
	configCmd.AddCommand(addCmd)
	configCmd.AddCommand(removeCmd)
}
