package cli

import (
	"mneme/internal/config"

	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Configuration commands",
	Long:  `Configuration commands`,
}

var showCmd = &cobra.Command{
	Use:   "show",
	Short: "show configuration values",
	Long:  `show configuration values`,
	Run:   config.ShowCmdExecute,
}

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "add path to index",
	Long:  `add path to index`,
	Run:   config.AddCmdExecute,
}

var removeCmd = &cobra.Command{
	Use:   "remove",
	Short: "remove path from indexing",
	Long:  `remove path from indexing`,
	Run:   config.RemoveCmdExecute,
}

func init() {
	removeCmd.Flags().BoolP("all", "a", false, "Remove all paths")

	configCmd.AddCommand(showCmd)
	configCmd.AddCommand(addCmd)
	configCmd.AddCommand(removeCmd)
}
