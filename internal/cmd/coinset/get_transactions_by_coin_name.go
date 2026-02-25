package cmd

import "github.com/spf13/cobra"

var getTransactionsByCoinNameIncludeTransactions bool

func init() {
	getTransactionsByCoinNameCmd.Flags().BoolVar(&getTransactionsByCoinNameIncludeTransactions, "include-transactions", false, "Include created/spent transaction payloads when available")
	rootCmd.AddCommand(getTransactionsByCoinNameCmd)
}

var getTransactionsByCoinNameCmd = &cobra.Command{
	Use:   "get_transactions_by_coin_name <coin_name>",
	Short: "Get associated transactions for a coin",
	Args: func(cmd *cobra.Command, args []string) error {
		if err := cobra.ExactArgs(1)(cmd, args); err != nil {
			return err
		}
		_, err := normalizeHexArg(args[0], "coin_name")
		return err
	},
	Run: func(cmd *cobra.Command, args []string) {
		coinName, _ := normalizeHexArg(args[0], "coin_name")
		jsonData := map[string]interface{}{
			"coin_name": coinName,
		}
		if getTransactionsByCoinNameIncludeTransactions {
			jsonData["include_transactions"] = true
		}
		makeRequest("get_transactions_by_coin_name", jsonData)
	},
}
