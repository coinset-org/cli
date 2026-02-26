package cmd

import "github.com/spf13/cobra"

var (
	getTransactionsByCatAssetIDCursor string
	getTransactionsByCatAssetIDLimit  int
	getTransactionsByCatAssetIDOrder  string
	getTransactionsByCatAssetIDPretty bool
)

func init() {
	getTransactionsByCatAssetIDCmd.Flags().StringVar(&getTransactionsByCatAssetIDCursor, "cursor", "", "Pagination cursor from a previous response")
	getTransactionsByCatAssetIDCmd.Flags().IntVar(&getTransactionsByCatAssetIDLimit, "limit", 0, "Maximum number of records to return")
	getTransactionsByCatAssetIDCmd.Flags().StringVar(&getTransactionsByCatAssetIDOrder, "order", "", `Result order: "asc" or "desc"`)
	getTransactionsByCatAssetIDCmd.Flags().BoolVarP(&getTransactionsByCatAssetIDPretty, "pretty", "p", false, "display formatted transaction output")
	rootCmd.AddCommand(getTransactionsByCatAssetIDCmd)
}

var getTransactionsByCatAssetIDCmd = &cobra.Command{
	Use:   "get_transactions_by_cat_asset_id <asset_id>",
	Short: "List transactions by CAT asset id",
	Args: func(cmd *cobra.Command, args []string) error {
		if err := cobra.ExactArgs(1)(cmd, args); err != nil {
			return err
		}
		if _, err := normalizeHexArg(args[0], "asset_id"); err != nil {
			return err
		}
		return applyPaginationOptions(map[string]interface{}{}, getTransactionsByCatAssetIDCursor, getTransactionsByCatAssetIDLimit, getTransactionsByCatAssetIDOrder)
	},
	Run: func(cmd *cobra.Command, args []string) {
		assetID, _ := normalizeHexArg(args[0], "asset_id")
		jsonData := map[string]interface{}{
			"asset_id": assetID,
		}
		_ = applyPaginationOptions(jsonData, getTransactionsByCatAssetIDCursor, getTransactionsByCatAssetIDLimit, getTransactionsByCatAssetIDOrder)
		makeTransactionRequest("get_transactions_by_cat_asset_id", jsonData, getTransactionsByCatAssetIDPretty)
	},
}
