package cmd

import "github.com/spf13/cobra"

var (
	getCatBalancesByP2AssetIDs []string
	getCatBalancesByP2Cursor   string
	getCatBalancesByP2Limit    int
	getCatBalancesByP2Order    string
)

func init() {
	getCatBalancesByP2Cmd.Flags().StringArrayVar(&getCatBalancesByP2AssetIDs, "asset-id", []string{}, "Optional CAT asset id filter (repeatable)")
	getCatBalancesByP2Cmd.Flags().StringVar(&getCatBalancesByP2Cursor, "cursor", "", "Pagination cursor from a previous response")
	getCatBalancesByP2Cmd.Flags().IntVar(&getCatBalancesByP2Limit, "limit", 0, "Maximum number of records to return")
	getCatBalancesByP2Cmd.Flags().StringVar(&getCatBalancesByP2Order, "order", "", `Result order: "asc" or "desc"`)
	rootCmd.AddCommand(getCatBalancesByP2Cmd)
}

var getCatBalancesByP2Cmd = &cobra.Command{
	Use:   "get_cat_balances_by_p2 <p2_or_address>",
	Short: "List CAT balances by p2",
	Args: func(cmd *cobra.Command, args []string) error {
		if err := cobra.ExactArgs(1)(cmd, args); err != nil {
			return err
		}
		if _, err := normalizeP2Arg(args[0]); err != nil {
			return err
		}

		for _, assetID := range getCatBalancesByP2AssetIDs {
			if _, err := normalizeHexArg(assetID, "asset_id"); err != nil {
				return err
			}
		}

		return applyPaginationOptions(map[string]interface{}{}, getCatBalancesByP2Cursor, getCatBalancesByP2Limit, getCatBalancesByP2Order)
	},
	Run: func(cmd *cobra.Command, args []string) {
		p2, _ := normalizeP2Arg(args[0])
		jsonData := map[string]interface{}{
			"p2": p2,
		}

		if len(getCatBalancesByP2AssetIDs) > 0 {
			normalized := make([]string, 0, len(getCatBalancesByP2AssetIDs))
			for _, assetID := range getCatBalancesByP2AssetIDs {
				formattedAssetID, _ := normalizeHexArg(assetID, "asset_id")
				normalized = append(normalized, formattedAssetID)
			}
			jsonData["asset_ids"] = normalized
		} else {
			_ = applyPaginationOptions(jsonData, getCatBalancesByP2Cursor, getCatBalancesByP2Limit, getCatBalancesByP2Order)
		}

		makeRequest("get_cat_balances_by_p2", jsonData)
	},
}
