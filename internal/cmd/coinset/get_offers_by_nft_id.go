package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	getOffersByNftIDStatus string
	getOffersByNftIDFilter string
	getOffersByNftIDCursor string
	getOffersByNftIDLimit  int
	getOffersByNftIDOrder  string
)

func init() {
	getOffersByNftIDCmd.Flags().StringVar(&getOffersByNftIDStatus, "status", "", "Lifecycle status: open,pending,confirmed,cancel_pending,cancelled,expired")
	getOffersByNftIDCmd.Flags().StringVar(&getOffersByNftIDFilter, "filter", "", "Offer side filter: all,requested,offered")
	getOffersByNftIDCmd.Flags().StringVar(&getOffersByNftIDCursor, "cursor", "", "Pagination cursor from a previous response")
	getOffersByNftIDCmd.Flags().IntVar(&getOffersByNftIDLimit, "limit", 0, "Maximum number of records to return")
	getOffersByNftIDCmd.Flags().StringVar(&getOffersByNftIDOrder, "order", "", `Result order: "asc" or "desc"`)
	rootCmd.AddCommand(getOffersByNftIDCmd)
}

var getOffersByNftIDCmd = &cobra.Command{
	Use:   "get_offers_by_nft_id <nft_id>",
	Short: "List offers by NFT id",
	Args: func(cmd *cobra.Command, args []string) error {
		if err := cobra.ExactArgs(1)(cmd, args); err != nil {
			return err
		}
		if args[0] == "" {
			return fmt.Errorf("nft_id is required")
		}
		if getOffersByNftIDStatus != "" && !isValidOfferStatus(getOffersByNftIDStatus) {
			return fmt.Errorf("invalid status: %s", getOffersByNftIDStatus)
		}
		if getOffersByNftIDFilter != "" && !isValidOfferFilter(getOffersByNftIDFilter) {
			return fmt.Errorf("invalid filter: %s", getOffersByNftIDFilter)
		}
		return applyPaginationOptions(map[string]interface{}{}, getOffersByNftIDCursor, getOffersByNftIDLimit, getOffersByNftIDOrder)
	},
	Run: func(cmd *cobra.Command, args []string) {
		jsonData := map[string]interface{}{
			"nft_id": normalizeMaybeHex(args[0]),
		}
		if getOffersByNftIDStatus != "" {
			jsonData["status"] = getOffersByNftIDStatus
		}
		if getOffersByNftIDFilter != "" {
			jsonData["filter"] = getOffersByNftIDFilter
		}
		_ = applyPaginationOptions(jsonData, getOffersByNftIDCursor, getOffersByNftIDLimit, getOffersByNftIDOrder)
		makeRequest("get_offers_by_nft_id", jsonData)
	},
}
