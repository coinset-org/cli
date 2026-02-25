package cmd

import "github.com/spf13/cobra"

var (
	getCatAssetsByP2IncludeSpentCoins bool
	getCatAssetsByP2Cursor            string
	getCatAssetsByP2Limit             int
	getCatAssetsByP2Order             string
)

func init() {
	getCatAssetsByP2Cmd.Flags().BoolVar(&getCatAssetsByP2IncludeSpentCoins, "include-spent-coins", false, "Include spent coins")
	getCatAssetsByP2Cmd.Flags().StringVar(&getCatAssetsByP2Cursor, "cursor", "", "Pagination cursor from a previous response")
	getCatAssetsByP2Cmd.Flags().IntVar(&getCatAssetsByP2Limit, "limit", 0, "Maximum number of records to return")
	getCatAssetsByP2Cmd.Flags().StringVar(&getCatAssetsByP2Order, "order", "", `Result order: "asc" or "desc"`)
	rootCmd.AddCommand(getCatAssetsByP2Cmd)
}

var getCatAssetsByP2Cmd = &cobra.Command{
	Use:   "get_cat_assets_by_p2 <p2_or_address>",
	Short: "List CAT asset ids by p2",
	Args: func(cmd *cobra.Command, args []string) error {
		if err := cobra.ExactArgs(1)(cmd, args); err != nil {
			return err
		}
		if _, err := normalizeP2Arg(args[0]); err != nil {
			return err
		}
		return applyPaginationOptions(map[string]interface{}{}, getCatAssetsByP2Cursor, getCatAssetsByP2Limit, getCatAssetsByP2Order)
	},
	Run: func(cmd *cobra.Command, args []string) {
		p2, _ := normalizeP2Arg(args[0])
		jsonData := map[string]interface{}{
			"p2": p2,
		}
		if getCatAssetsByP2IncludeSpentCoins {
			jsonData["include_spent_coins"] = true
		}
		_ = applyPaginationOptions(jsonData, getCatAssetsByP2Cursor, getCatAssetsByP2Limit, getCatAssetsByP2Order)
		makeRequest("get_cat_assets_by_p2", jsonData)
	},
}
