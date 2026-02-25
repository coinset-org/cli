package cmd

import "github.com/spf13/cobra"

var (
	getTransactionsByP2Cursor string
	getTransactionsByP2Limit  int
	getTransactionsByP2Order  string
)

func init() {
	getTransactionsByP2Cmd.Flags().StringVar(&getTransactionsByP2Cursor, "cursor", "", "Pagination cursor from a previous response")
	getTransactionsByP2Cmd.Flags().IntVar(&getTransactionsByP2Limit, "limit", 0, "Maximum number of records to return")
	getTransactionsByP2Cmd.Flags().StringVar(&getTransactionsByP2Order, "order", "", `Result order: "asc" or "desc"`)
	rootCmd.AddCommand(getTransactionsByP2Cmd)
}

var getTransactionsByP2Cmd = &cobra.Command{
	Use:   "get_transactions_by_p2 <p2_or_address>",
	Short: "List confirmed transactions by p2",
	Args: func(cmd *cobra.Command, args []string) error {
		if err := cobra.ExactArgs(1)(cmd, args); err != nil {
			return err
		}
		if _, err := normalizeP2Arg(args[0]); err != nil {
			return err
		}
		return applyPaginationOptions(map[string]interface{}{}, getTransactionsByP2Cursor, getTransactionsByP2Limit, getTransactionsByP2Order)
	},
	Run: func(cmd *cobra.Command, args []string) {
		p2, _ := normalizeP2Arg(args[0])
		jsonData := map[string]interface{}{
			"p2": p2,
		}
		_ = applyPaginationOptions(jsonData, getTransactionsByP2Cursor, getTransactionsByP2Limit, getTransactionsByP2Order)
		makeRequest("get_transactions_by_p2", jsonData)
	},
}
