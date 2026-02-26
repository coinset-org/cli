package cmd

import "github.com/spf13/cobra"

var (
	getPendingTransactionsByPuzzleHashCursor string
	getPendingTransactionsByPuzzleHashLimit  int
	getPendingTransactionsByPuzzleHashOrder  string
	getPendingTransactionsByPuzzleHashPretty bool
)

func init() {
	getPendingTransactionsByPuzzleHashCmd.Flags().StringVar(&getPendingTransactionsByPuzzleHashCursor, "cursor", "", "Pagination cursor from a previous response")
	getPendingTransactionsByPuzzleHashCmd.Flags().IntVar(&getPendingTransactionsByPuzzleHashLimit, "limit", 0, "Maximum number of records to return")
	getPendingTransactionsByPuzzleHashCmd.Flags().StringVar(&getPendingTransactionsByPuzzleHashOrder, "order", "", `Result order: "asc" or "desc"`)
	getPendingTransactionsByPuzzleHashCmd.Flags().BoolVarP(&getPendingTransactionsByPuzzleHashPretty, "pretty", "p", false, "display formatted transaction output")
	rootCmd.AddCommand(getPendingTransactionsByPuzzleHashCmd)
}

var getPendingTransactionsByPuzzleHashCmd = &cobra.Command{
	Use:   "get_pending_transactions_by_puzzle_hash <puzzle_hash_or_address>",
	Short: "Back-compat alias for get_pending_transactions_by_p2",
	Args: func(cmd *cobra.Command, args []string) error {
		if err := cobra.ExactArgs(1)(cmd, args); err != nil {
			return err
		}
		if _, err := normalizeP2Arg(args[0]); err != nil {
			return err
		}
		return applyPaginationOptions(map[string]interface{}{}, getPendingTransactionsByPuzzleHashCursor, getPendingTransactionsByPuzzleHashLimit, getPendingTransactionsByPuzzleHashOrder)
	},
	Run: func(cmd *cobra.Command, args []string) {
		p2, _ := normalizeP2Arg(args[0])
		jsonData := map[string]interface{}{
			"p2": p2,
		}
		_ = applyPaginationOptions(jsonData, getPendingTransactionsByPuzzleHashCursor, getPendingTransactionsByPuzzleHashLimit, getPendingTransactionsByPuzzleHashOrder)
		makeTransactionRequest("get_pending_transactions_by_puzzle_hash", jsonData, getPendingTransactionsByPuzzleHashPretty)
	},
}
