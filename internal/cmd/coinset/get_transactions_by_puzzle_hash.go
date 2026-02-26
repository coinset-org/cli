package cmd

import "github.com/spf13/cobra"

var (
	getTransactionsByPuzzleHashCursor string
	getTransactionsByPuzzleHashLimit  int
	getTransactionsByPuzzleHashOrder  string
	getTransactionsByPuzzleHashPretty bool
)

func init() {
	getTransactionsByPuzzleHashCmd.Flags().StringVar(&getTransactionsByPuzzleHashCursor, "cursor", "", "Pagination cursor from a previous response")
	getTransactionsByPuzzleHashCmd.Flags().IntVar(&getTransactionsByPuzzleHashLimit, "limit", 0, "Maximum number of records to return")
	getTransactionsByPuzzleHashCmd.Flags().StringVar(&getTransactionsByPuzzleHashOrder, "order", "", `Result order: "asc" or "desc"`)
	getTransactionsByPuzzleHashCmd.Flags().BoolVarP(&getTransactionsByPuzzleHashPretty, "pretty", "p", false, "display formatted transaction output")
	rootCmd.AddCommand(getTransactionsByPuzzleHashCmd)
}

var getTransactionsByPuzzleHashCmd = &cobra.Command{
	Use:   "get_transactions_by_puzzle_hash <puzzle_hash_or_address>",
	Short: "Back-compat alias for get_transactions_by_p2",
	Args: func(cmd *cobra.Command, args []string) error {
		if err := cobra.ExactArgs(1)(cmd, args); err != nil {
			return err
		}
		if _, err := normalizeP2Arg(args[0]); err != nil {
			return err
		}
		return applyPaginationOptions(map[string]interface{}{}, getTransactionsByPuzzleHashCursor, getTransactionsByPuzzleHashLimit, getTransactionsByPuzzleHashOrder)
	},
	Run: func(cmd *cobra.Command, args []string) {
		p2, _ := normalizeP2Arg(args[0])
		jsonData := map[string]interface{}{
			"p2": p2,
		}
		_ = applyPaginationOptions(jsonData, getTransactionsByPuzzleHashCursor, getTransactionsByPuzzleHashLimit, getTransactionsByPuzzleHashOrder)
		makeTransactionRequest("get_transactions_by_puzzle_hash", jsonData, getTransactionsByPuzzleHashPretty)
	},
}
