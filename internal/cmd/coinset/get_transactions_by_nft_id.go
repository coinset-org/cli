package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	getTransactionsByNftIDCursor string
	getTransactionsByNftIDLimit  int
	getTransactionsByNftIDOrder  string
	getTransactionsByNftIDPretty bool
)

func init() {
	getTransactionsByNftIDCmd.Flags().StringVar(&getTransactionsByNftIDCursor, "cursor", "", "Pagination cursor from a previous response")
	getTransactionsByNftIDCmd.Flags().IntVar(&getTransactionsByNftIDLimit, "limit", 0, "Maximum number of records to return")
	getTransactionsByNftIDCmd.Flags().StringVar(&getTransactionsByNftIDOrder, "order", "", `Result order: "asc" or "desc"`)
	getTransactionsByNftIDCmd.Flags().BoolVarP(&getTransactionsByNftIDPretty, "pretty", "p", false, "display formatted transaction output")
	rootCmd.AddCommand(getTransactionsByNftIDCmd)
}

var getTransactionsByNftIDCmd = &cobra.Command{
	Use:   "get_transactions_by_nft_id <nft_id>",
	Short: "List transactions by NFT id",
	Args: func(cmd *cobra.Command, args []string) error {
		if err := cobra.ExactArgs(1)(cmd, args); err != nil {
			return err
		}
		if args[0] == "" {
			return fmt.Errorf("nft_id is required")
		}
		return applyPaginationOptions(map[string]interface{}{}, getTransactionsByNftIDCursor, getTransactionsByNftIDLimit, getTransactionsByNftIDOrder)
	},
	Run: func(cmd *cobra.Command, args []string) {
		jsonData := map[string]interface{}{
			"nft_id": normalizeMaybeHex(args[0]),
		}
		_ = applyPaginationOptions(jsonData, getTransactionsByNftIDCursor, getTransactionsByNftIDLimit, getTransactionsByNftIDOrder)
		makeTransactionRequest("get_transactions_by_nft_id", jsonData, getTransactionsByNftIDPretty)
	},
}
