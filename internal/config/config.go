package config

import "github.com/spf13/cobra"

var ConfigCmd = &cobra.Command{
	Use:   "config",
	Short: "Configuration commands",
	Long:  `Configuration commands`,
}

var showCmd = &cobra.Command{
	Use:   "show",
	Short: "show configuration values",
	Long:  `show configuration values`,
	Run:   showCmdExecute,
}

func init() {
	ConfigCmd.AddCommand(showCmd)
}
