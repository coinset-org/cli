package cmd

import "github.com/spf13/cobra"

func init() {
	rootCmd.AddCommand(getTransactionCmd)
}

var getTransactionCmd = &cobra.Command{
	Use:   "get_transaction <tx_id>",
	Short: "Get a transaction by tx id",
	Args: func(cmd *cobra.Command, args []string) error {
		if err := cobra.ExactArgs(1)(cmd, args); err != nil {
			return err
		}
		_, err := normalizeHexArg(args[0], "tx_id")
		return err
	},
	Run: func(cmd *cobra.Command, args []string) {
		txID, _ := normalizeHexArg(args[0], "tx_id")
		makeRequest("get_transaction", map[string]interface{}{
			"tx_id": txID,
		})
	},
}
