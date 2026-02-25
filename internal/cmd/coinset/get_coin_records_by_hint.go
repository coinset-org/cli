package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	crByHintCmdIncludeSpentCoins bool
	crByHintCmdStart             int
	crByHintCmdEnd               int
	crByHintCmdCursor            string
	crByHintCmdLimit             int
	crByHintCmdOrder             string
)

func init() {
	rootCmd.AddCommand(getCoinRecordsByHintCmd)

	// Define flags for the optional arguments
	getCoinRecordsByHintCmd.Flags().BoolVarP(&crByHintCmdIncludeSpentCoins, "include-spent-coins", "s", false, "Include spent coins")
	getCoinRecordsByHintCmd.Flags().IntVarP(&crByHintCmdStart, "start-height", "", -1, "Start height")
	getCoinRecordsByHintCmd.Flags().IntVarP(&crByHintCmdEnd, "end-height", "", -1, "End height")
	getCoinRecordsByHintCmd.Flags().StringVar(&crByHintCmdCursor, "cursor", "", "Pagination cursor from a previous truncated response")
	getCoinRecordsByHintCmd.Flags().IntVar(&crByHintCmdLimit, "limit", 0, "Maximum number of records to return")
	getCoinRecordsByHintCmd.Flags().StringVar(&crByHintCmdOrder, "order", "", `Result order: "asc" or "desc"`)
}

var getCoinRecordsByHintCmd = &cobra.Command{
	Use: "get_coin_records_by_hint <hint>",
	Args: func(cmd *cobra.Command, args []string) error {
		if err := cobra.ExactArgs(1)(cmd, args); err != nil {
			return err
		}
		if isHex(args[0]) {
			if crByHintCmdLimit < 0 {
				return fmt.Errorf("limit must be >= 1 when provided")
			}
			if crByHintCmdOrder != "" && !isAscOrDesc(crByHintCmdOrder) {
				return fmt.Errorf(`invalid order %q: must be "asc" or "desc"`, crByHintCmdOrder)
			}
			return nil
		}
		return fmt.Errorf("invalid hex value specified: %s", args[0])
	},
	Short: "Retrieves coin records by hint",
	Long:  "Retrieves coin records by hint",
	Run: func(cmd *cobra.Command, args []string) {
		jsonData := map[string]interface{}{}

		jsonData["hint"] = args[0]

		if crByHintCmdIncludeSpentCoins {
			jsonData["include_spent_coins"] = true
		}
		if crByHintCmdStart != -1 {
			jsonData["start_height"] = crByHintCmdStart
		}
		if crByHintCmdEnd != -1 {
			jsonData["end_height"] = crByHintCmdEnd
		}
		if crByHintCmdCursor != "" {
			jsonData["cursor"] = crByHintCmdCursor
		}
		if crByHintCmdLimit > 0 {
			jsonData["limit"] = crByHintCmdLimit
		}
		if crByHintCmdOrder != "" {
			jsonData["order"] = crByHintCmdOrder
		}
		makeRequest("get_coin_records_by_hint", jsonData)
	},
}
