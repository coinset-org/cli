package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	getOffersByP2Status string
	getOffersByP2Cursor string
	getOffersByP2Limit  int
	getOffersByP2Order  string
)

func init() {
	getOffersByP2Cmd.Flags().StringVar(&getOffersByP2Status, "status", "", "Lifecycle status: open,pending,confirmed,cancel_pending,cancelled,expired")
	getOffersByP2Cmd.Flags().StringVar(&getOffersByP2Cursor, "cursor", "", "Pagination cursor from a previous response")
	getOffersByP2Cmd.Flags().IntVar(&getOffersByP2Limit, "limit", 0, "Maximum number of records to return")
	getOffersByP2Cmd.Flags().StringVar(&getOffersByP2Order, "order", "", `Result order: "asc" or "desc"`)
	rootCmd.AddCommand(getOffersByP2Cmd)
}

var getOffersByP2Cmd = &cobra.Command{
	Use:   "get_offers_by_p2 <p2_or_address>",
	Short: "List offers by maker p2",
	Args: func(cmd *cobra.Command, args []string) error {
		if err := cobra.ExactArgs(1)(cmd, args); err != nil {
			return err
		}
		if _, err := normalizeP2Arg(args[0]); err != nil {
			return err
		}
		if getOffersByP2Status != "" && !isValidOfferStatus(getOffersByP2Status) {
			return fmt.Errorf("invalid status: %s", getOffersByP2Status)
		}
		return applyPaginationOptions(map[string]interface{}{}, getOffersByP2Cursor, getOffersByP2Limit, getOffersByP2Order)
	},
	Run: func(cmd *cobra.Command, args []string) {
		p2, _ := normalizeP2Arg(args[0])
		jsonData := map[string]interface{}{
			"p2": p2,
		}
		if getOffersByP2Status != "" {
			jsonData["status"] = getOffersByP2Status
		}
		_ = applyPaginationOptions(jsonData, getOffersByP2Cursor, getOffersByP2Limit, getOffersByP2Order)
		makeRequest("get_offers_by_p2", jsonData)
	},
}
