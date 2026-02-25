package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	crByHintsIncludeSpentCoins bool
	crByHintsStart             int
	crByHintsEnd               int
	crByHintsCursor            string
	crByHintsLimit             int
	crByHintsOrder             string
)

func init() {
	getCoinRecordsByHintsCmd.Flags().BoolVarP(&crByHintsIncludeSpentCoins, "include-spent-coins", "s", false, "Include spent coins")
	getCoinRecordsByHintsCmd.Flags().IntVarP(&crByHintsStart, "start-height", "", -1, "Start height")
	getCoinRecordsByHintsCmd.Flags().IntVarP(&crByHintsEnd, "end-height", "", -1, "End height")
	getCoinRecordsByHintsCmd.Flags().StringVar(&crByHintsCursor, "cursor", "", "Pagination cursor from a previous truncated response")
	getCoinRecordsByHintsCmd.Flags().IntVar(&crByHintsLimit, "limit", 0, "Maximum number of records to return")
	getCoinRecordsByHintsCmd.Flags().StringVar(&crByHintsOrder, "order", "", `Result order: "asc" or "desc"`)
	rootCmd.AddCommand(getCoinRecordsByHintsCmd)
}

var getCoinRecordsByHintsCmd = &cobra.Command{
	Use: "get_coin_records_by_hints <hint1> <hint2> ...",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return fmt.Errorf("at least one hint is required")
		}
		for _, hint := range args {
			if !isHex(hint) {
				return fmt.Errorf("invalid hex value specified: %s", hint)
			}
		}
		if crByHintsLimit < 0 {
			return fmt.Errorf("limit must be >= 1 when provided")
		}
		if crByHintsOrder != "" && !isAscOrDesc(crByHintsOrder) {
			return fmt.Errorf(`invalid order %q: must be "asc" or "desc"`, crByHintsOrder)
		}
		return nil
	},
	Short: "Retrieves coin records by multiple hints",
	Long:  "Retrieves coin records by multiple hints",
	Run: func(cmd *cobra.Command, args []string) {
		var hints []string
		for _, hint := range args {
			hints = append(hints, formatHex(hint))
		}
		jsonData := map[string]interface{}{
			"hints": hints,
		}
		if crByHintsIncludeSpentCoins {
			jsonData["include_spent_coins"] = true
		}
		if crByHintsStart != -1 {
			jsonData["start_height"] = crByHintsStart
		}
		if crByHintsEnd != -1 {
			jsonData["end_height"] = crByHintsEnd
		}
		if crByHintsCursor != "" {
			jsonData["cursor"] = crByHintsCursor
		}
		if crByHintsLimit > 0 {
			jsonData["limit"] = crByHintsLimit
		}
		if crByHintsOrder != "" {
			jsonData["order"] = crByHintsOrder
		}
		makeRequest("get_coin_records_by_hints", jsonData)
	},
}
