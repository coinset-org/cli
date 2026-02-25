package cmd

import "github.com/spf13/cobra"

func init() {
	rootCmd.AddCommand(getRawTransactionByIDCmd)
}

var getRawTransactionByIDCmd = &cobra.Command{
	Use:   "get_raw_transaction_by_id <tx_id>",
	Short: "Get a raw mempool-style transaction item by id",
	Args: func(cmd *cobra.Command, args []string) error {
		if err := cobra.ExactArgs(1)(cmd, args); err != nil {
			return err
		}
		_, err := normalizeHexArg(args[0], "tx_id")
		return err
	},
	Run: func(cmd *cobra.Command, args []string) {
		txID, _ := normalizeHexArg(args[0], "tx_id")
		makeRequest("get_raw_transaction_by_id", map[string]interface{}{
			"tx_id": txID,
		})
	},
}
