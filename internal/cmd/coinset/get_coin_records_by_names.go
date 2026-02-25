package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	crByNamesIncludeSpentCoins bool
	crByNamesStart             int
	crByNamesEnd               int
	crByNamesCursor            string
	crByNamesLimit             int
	crByNamesOrder             string
)

func init() {
	getCoinRecordsByNamesCmd.Flags().BoolVarP(&crByNamesIncludeSpentCoins, "include-spent-coins", "s", false, "Include spent coins")
	getCoinRecordsByNamesCmd.Flags().IntVarP(&crByNamesStart, "start", "", -1, "Start height")
	getCoinRecordsByNamesCmd.Flags().IntVarP(&crByNamesEnd, "end", "", -1, "End height")
	getCoinRecordsByNamesCmd.Flags().StringVar(&crByNamesCursor, "cursor", "", "Pagination cursor from a previous truncated response")
	getCoinRecordsByNamesCmd.Flags().IntVar(&crByNamesLimit, "limit", 0, "Maximum number of records to return")
	getCoinRecordsByNamesCmd.Flags().StringVar(&crByNamesOrder, "order", "", `Result order: "asc" or "desc"`)
	rootCmd.AddCommand(getCoinRecordsByNamesCmd)
}

var getCoinRecordsByNamesCmd = &cobra.Command{
	Use: "get_coin_records_by_names <name1> <name2> ...",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return fmt.Errorf("at least one name is required")
		}
		for _, name := range args {
			if !isHex(name) {
				return fmt.Errorf("invalid hex value specified: %s", name)
			}
		}
		if crByNamesLimit < 0 {
			return fmt.Errorf("limit must be >= 1 when provided")
		}
		if crByNamesOrder != "" && !isAscOrDesc(crByNamesOrder) {
			return fmt.Errorf(`invalid order %q: must be "asc" or "desc"`, crByNamesOrder)
		}
		return nil
	},
	Short: "Retrieves coin records by their names",
	Long:  "Retrieves coin records by their names",
	Run: func(cmd *cobra.Command, args []string) {
		var names []string
		for _, name := range args {
			names = append(names, formatHex(name))
		}
		jsonData := map[string]interface{}{
			"names": names,
		}
		if crByNamesIncludeSpentCoins {
			jsonData["include_spent_coins"] = true
		}
		if crByNamesStart != -1 {
			jsonData["start_height"] = crByNamesStart
		}
		if crByNamesEnd != -1 {
			jsonData["end_height"] = crByNamesEnd
		}
		if crByNamesCursor != "" {
			jsonData["cursor"] = crByNamesCursor
		}
		if crByNamesLimit > 0 {
			jsonData["limit"] = crByNamesLimit
		}
		if crByNamesOrder != "" {
			jsonData["order"] = crByNamesOrder
		}
		makeRequest("get_coin_records_by_names", jsonData)
	},
}
