package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	crByParentIdsIncludeSpentCoins bool
	crByParentIdsStart             int
	crByParentIdsEnd               int
	crByParentIdsCursor            string
	crByParentIdsLimit             int
	crByParentIdsOrder             string
)

func init() {
	rootCmd.AddCommand(getCoinRecordsByParentIdsCmd)

	// Define flags for the optional arguments
	getCoinRecordsByParentIdsCmd.Flags().BoolVarP(&crByParentIdsIncludeSpentCoins, "include-spent-coins", "s", false, "Include spent coins")
	getCoinRecordsByParentIdsCmd.Flags().IntVarP(&crByParentIdsStart, "start-height", "", -1, "Start height")
	getCoinRecordsByParentIdsCmd.Flags().IntVarP(&crByParentIdsEnd, "end-height", "", -1, "End height")
	getCoinRecordsByParentIdsCmd.Flags().StringVar(&crByParentIdsCursor, "cursor", "", "Pagination cursor from a previous truncated response")
	getCoinRecordsByParentIdsCmd.Flags().IntVar(&crByParentIdsLimit, "limit", 0, "Maximum number of records to return")
	getCoinRecordsByParentIdsCmd.Flags().StringVar(&crByParentIdsOrder, "order", "", `Result order: "asc" or "desc"`)
}

var getCoinRecordsByParentIdsCmd = &cobra.Command{
	Use: "get_coin_records_by_parent_ids <parent_id> <parent_id> ...",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return fmt.Errorf("at least one parent ID is required")
		}
		for _, name := range args {
			if !isHex(name) {
				return fmt.Errorf("invalid hex value specified: %s", name)
			}
		}
		if crByParentIdsLimit < 0 {
			return fmt.Errorf("limit must be >= 1 when provided")
		}
		if crByParentIdsOrder != "" && !isAscOrDesc(crByParentIdsOrder) {
			return fmt.Errorf(`invalid order %q: must be "asc" or "desc"`, crByParentIdsOrder)
		}
		return nil
	},
	Short: "Retrieves coin records by parent IDs",
	Long:  "Retrieves coin records by parent IDs",
	Run: func(cmd *cobra.Command, args []string) {
		var parentIds []string
		for _, parentId := range args {
			parentIds = append(parentIds, formatHex(parentId))
		}
		jsonData := map[string]interface{}{}
		jsonData["parent_ids"] = parentIds
		if crByParentIdsIncludeSpentCoins {
			jsonData["include_spent_coins"] = true
		}
		if crByParentIdsStart != -1 {
			jsonData["start_height"] = crByParentIdsStart
		}
		if crByParentIdsEnd != -1 {
			jsonData["end_height"] = crByParentIdsEnd
		}
		if crByParentIdsCursor != "" {
			jsonData["cursor"] = crByParentIdsCursor
		}
		if crByParentIdsLimit > 0 {
			jsonData["limit"] = crByParentIdsLimit
		}
		if crByParentIdsOrder != "" {
			jsonData["order"] = crByParentIdsOrder
		}
		makeRequest("get_coin_records_by_parent_ids", jsonData)
	},
}
