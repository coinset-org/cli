package cmd

import "github.com/spf13/cobra"

func init() {
	rootCmd.AddCommand(getSingletonInfoCmd)
}

var getSingletonInfoCmd = &cobra.Command{
	Use:   "get_singleton_info <launcher_id>",
	Short: "Get singleton info by launcher id",
	Args: func(cmd *cobra.Command, args []string) error {
		if err := cobra.ExactArgs(1)(cmd, args); err != nil {
			return err
		}
		_, err := normalizeHexArg(args[0], "launcher_id")
		return err
	},
	Run: func(cmd *cobra.Command, args []string) {
		launcherID, _ := normalizeHexArg(args[0], "launcher_id")
		makeRequest("get_singleton_info", map[string]interface{}{
			"launcher_id": launcherID,
		})
	},
}
