package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	getOffersByCatAssetIDStatus string
	getOffersByCatAssetIDFilter string
	getOffersByCatAssetIDCursor string
	getOffersByCatAssetIDLimit  int
	getOffersByCatAssetIDOrder  string
)

func init() {
	getOffersByCatAssetIDCmd.Flags().StringVar(&getOffersByCatAssetIDStatus, "status", "", "Lifecycle status: open,pending,confirmed,cancel_pending,cancelled,expired")
	getOffersByCatAssetIDCmd.Flags().StringVar(&getOffersByCatAssetIDFilter, "filter", "", "Offer side filter: all,requested,offered")
	getOffersByCatAssetIDCmd.Flags().StringVar(&getOffersByCatAssetIDCursor, "cursor", "", "Pagination cursor from a previous response")
	getOffersByCatAssetIDCmd.Flags().IntVar(&getOffersByCatAssetIDLimit, "limit", 0, "Maximum number of records to return")
	getOffersByCatAssetIDCmd.Flags().StringVar(&getOffersByCatAssetIDOrder, "order", "", `Result order: "asc" or "desc"`)
	rootCmd.AddCommand(getOffersByCatAssetIDCmd)
}

var getOffersByCatAssetIDCmd = &cobra.Command{
	Use:   "get_offers_by_cat_asset_id <asset_id>",
	Short: "List offers by CAT asset id",
	Args: func(cmd *cobra.Command, args []string) error {
		if err := cobra.ExactArgs(1)(cmd, args); err != nil {
			return err
		}
		if _, err := normalizeHexArg(args[0], "asset_id"); err != nil {
			return err
		}
		if getOffersByCatAssetIDStatus != "" && !isValidOfferStatus(getOffersByCatAssetIDStatus) {
			return fmt.Errorf("invalid status: %s", getOffersByCatAssetIDStatus)
		}
		if getOffersByCatAssetIDFilter != "" && !isValidOfferFilter(getOffersByCatAssetIDFilter) {
			return fmt.Errorf("invalid filter: %s", getOffersByCatAssetIDFilter)
		}
		return applyPaginationOptions(map[string]interface{}{}, getOffersByCatAssetIDCursor, getOffersByCatAssetIDLimit, getOffersByCatAssetIDOrder)
	},
	Run: func(cmd *cobra.Command, args []string) {
		assetID, _ := normalizeHexArg(args[0], "asset_id")
		jsonData := map[string]interface{}{
			"asset_id": assetID,
		}
		if getOffersByCatAssetIDStatus != "" {
			jsonData["status"] = getOffersByCatAssetIDStatus
		}
		if getOffersByCatAssetIDFilter != "" {
			jsonData["filter"] = getOffersByCatAssetIDFilter
		}
		_ = applyPaginationOptions(jsonData, getOffersByCatAssetIDCursor, getOffersByCatAssetIDLimit, getOffersByCatAssetIDOrder)
		makeRequest("get_offers_by_cat_asset_id", jsonData)
	},
}
