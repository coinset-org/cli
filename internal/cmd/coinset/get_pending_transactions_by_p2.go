package cmd

import "github.com/spf13/cobra"

var (
	getPendingTransactionsByP2Cursor string
	getPendingTransactionsByP2Limit  int
	getPendingTransactionsByP2Order  string
)

func init() {
	getPendingTransactionsByP2Cmd.Flags().StringVar(&getPendingTransactionsByP2Cursor, "cursor", "", "Pagination cursor from a previous response")
	getPendingTransactionsByP2Cmd.Flags().IntVar(&getPendingTransactionsByP2Limit, "limit", 0, "Maximum number of records to return")
	getPendingTransactionsByP2Cmd.Flags().StringVar(&getPendingTransactionsByP2Order, "order", "", `Result order: "asc" or "desc"`)
	rootCmd.AddCommand(getPendingTransactionsByP2Cmd)
}

var getPendingTransactionsByP2Cmd = &cobra.Command{
	Use:   "get_pending_transactions_by_p2 <p2_or_address>",
	Short: "List pending transactions by p2",
	Args: func(cmd *cobra.Command, args []string) error {
		if err := cobra.ExactArgs(1)(cmd, args); err != nil {
			return err
		}
		if _, err := normalizeP2Arg(args[0]); err != nil {
			return err
		}
		return applyPaginationOptions(map[string]interface{}{}, getPendingTransactionsByP2Cursor, getPendingTransactionsByP2Limit, getPendingTransactionsByP2Order)
	},
	Run: func(cmd *cobra.Command, args []string) {
		p2, _ := normalizeP2Arg(args[0])
		jsonData := map[string]interface{}{
			"p2": p2,
		}
		_ = applyPaginationOptions(jsonData, getPendingTransactionsByP2Cursor, getPendingTransactionsByP2Limit, getPendingTransactionsByP2Order)
		makeRequest("get_pending_transactions_by_p2", jsonData)
	},
}
