package cmd

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
)

var (
	getBlockTransactionsCursor string
	getBlockTransactionsLimit  int
	getBlockTransactionsOrder  string
)

func init() {
	getBlockTransactionsCmd.Flags().StringVar(&getBlockTransactionsCursor, "cursor", "", "Pagination cursor from a previous response")
	getBlockTransactionsCmd.Flags().IntVar(&getBlockTransactionsLimit, "limit", 0, "Maximum number of records to return")
	getBlockTransactionsCmd.Flags().StringVar(&getBlockTransactionsOrder, "order", "", `Result order: "asc" or "desc"`)
	rootCmd.AddCommand(getBlockTransactionsCmd)
}

var getBlockTransactionsCmd = &cobra.Command{
	Use:   "get_block_transactions <height_or_header_hash>",
	Short: "List transactions for a block",
	Args: func(cmd *cobra.Command, args []string) error {
		if err := cobra.ExactArgs(1)(cmd, args); err != nil {
			return err
		}

		if _, err := strconv.Atoi(args[0]); err != nil && !isHex(args[0]) {
			return fmt.Errorf("invalid block selector: %s", args[0])
		}

		return applyPaginationOptions(map[string]interface{}{}, getBlockTransactionsCursor, getBlockTransactionsLimit, getBlockTransactionsOrder)
	},
	Run: func(cmd *cobra.Command, args []string) {
		jsonData := map[string]interface{}{}
		if height, err := strconv.Atoi(args[0]); err == nil {
			jsonData["height"] = height
		} else {
			jsonData["header_hash"] = formatHex(args[0])
		}

		_ = applyPaginationOptions(jsonData, getBlockTransactionsCursor, getBlockTransactionsLimit, getBlockTransactionsOrder)
		makeRequest("get_block_transactions", jsonData)
	},
}
