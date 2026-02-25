package cmd

import "github.com/spf13/cobra"

var (
	getCatCoinsByP2AndAssetIDIncludeSpentCoins bool
	getCatCoinsByP2AndAssetIDCursor            string
	getCatCoinsByP2AndAssetIDLimit             int
	getCatCoinsByP2AndAssetIDOrder             string
)

func init() {
	getCatCoinsByP2AndAssetIDCmd.Flags().BoolVar(&getCatCoinsByP2AndAssetIDIncludeSpentCoins, "include-spent-coins", false, "Include spent coins")
	getCatCoinsByP2AndAssetIDCmd.Flags().StringVar(&getCatCoinsByP2AndAssetIDCursor, "cursor", "", "Pagination cursor from a previous response")
	getCatCoinsByP2AndAssetIDCmd.Flags().IntVar(&getCatCoinsByP2AndAssetIDLimit, "limit", 0, "Maximum number of records to return")
	getCatCoinsByP2AndAssetIDCmd.Flags().StringVar(&getCatCoinsByP2AndAssetIDOrder, "order", "", `Result order: "asc" or "desc"`)
	rootCmd.AddCommand(getCatCoinsByP2AndAssetIDCmd)
}

var getCatCoinsByP2AndAssetIDCmd = &cobra.Command{
	Use:   "get_cat_coins_by_p2_and_asset_id <p2_or_address> <asset_id>",
	Short: "List CAT coins by p2 and asset id",
	Args: func(cmd *cobra.Command, args []string) error {
		if err := cobra.ExactArgs(2)(cmd, args); err != nil {
			return err
		}
		if _, err := normalizeP2Arg(args[0]); err != nil {
			return err
		}
		if _, err := normalizeHexArg(args[1], "asset_id"); err != nil {
			return err
		}
		return applyPaginationOptions(map[string]interface{}{}, getCatCoinsByP2AndAssetIDCursor, getCatCoinsByP2AndAssetIDLimit, getCatCoinsByP2AndAssetIDOrder)
	},
	Run: func(cmd *cobra.Command, args []string) {
		p2, _ := normalizeP2Arg(args[0])
		assetID, _ := normalizeHexArg(args[1], "asset_id")
		jsonData := map[string]interface{}{
			"p2":       p2,
			"asset_id": assetID,
		}
		if getCatCoinsByP2AndAssetIDIncludeSpentCoins {
			jsonData["include_spent_coins"] = true
		}
		_ = applyPaginationOptions(jsonData, getCatCoinsByP2AndAssetIDCursor, getCatCoinsByP2AndAssetIDLimit, getCatCoinsByP2AndAssetIDOrder)
		makeRequest("get_cat_coins_by_p2_and_asset_id", jsonData)
	},
}
