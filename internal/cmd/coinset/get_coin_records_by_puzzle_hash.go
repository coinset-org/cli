package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	crByPuzzleHashIncludeSpentCoins bool
	crByPuzzleHashStart             int
	crByPuzzleHashEnd               int
	crByPuzzleHashCursor            string
	crByPuzzleHashLimit             int
	crByPuzzleHashOrder             string
)

func init() {
	getCoinRecordsByPuzzleHashCmd.Flags().BoolVarP(&crByPuzzleHashIncludeSpentCoins, "include-spent-coins", "s", false, "Include spent coins")
	getCoinRecordsByPuzzleHashCmd.Flags().IntVarP(&crByPuzzleHashStart, "start-height", "", -1, "Start height")
	getCoinRecordsByPuzzleHashCmd.Flags().IntVarP(&crByPuzzleHashEnd, "end-height", "", -1, "End height")
	getCoinRecordsByPuzzleHashCmd.Flags().StringVar(&crByPuzzleHashCursor, "cursor", "", "Pagination cursor from a previous truncated response")
	getCoinRecordsByPuzzleHashCmd.Flags().IntVar(&crByPuzzleHashLimit, "limit", 0, "Maximum number of records to return")
	getCoinRecordsByPuzzleHashCmd.Flags().StringVar(&crByPuzzleHashOrder, "order", "", `Result order: "asc" or "desc"`)
	rootCmd.AddCommand(getCoinRecordsByPuzzleHashCmd)
}

var getCoinRecordsByPuzzleHashCmd = &cobra.Command{
	Use: "get_coin_records_by_puzzle_hash <puzzle_hash_or_address>",
	Args: func(cmd *cobra.Command, args []string) error {
		if err := cobra.ExactArgs(1)(cmd, args); err != nil {
			return err
		}
		_, err := convertAddressOrPuzzleHash(args[0])
		if err != nil {
			return err
		}
		if crByPuzzleHashLimit < 0 {
			return fmt.Errorf("limit must be >= 1 when provided")
		}
		if crByPuzzleHashOrder != "" && !isAscOrDesc(crByPuzzleHashOrder) {
			return fmt.Errorf(`invalid order %q: must be "asc" or "desc"`, crByPuzzleHashOrder)
		}
		return nil
	},
	Short: "Retrieves coin records by their puzzle hash or address",
	Long:  "Retrieves coin records by their puzzle hash or address",
	Run: func(cmd *cobra.Command, args []string) {
		puzzleHash, _ := convertAddressOrPuzzleHash(args[0])
		jsonData := map[string]interface{}{}
		jsonData["puzzle_hash"] = puzzleHash
		if crByPuzzleHashIncludeSpentCoins {
			jsonData["include_spent_coins"] = true
		}
		if crByPuzzleHashStart != -1 {
			jsonData["start_height"] = crByPuzzleHashStart
		}
		if crByPuzzleHashEnd != -1 {
			jsonData["end_height"] = crByPuzzleHashEnd
		}
		if crByPuzzleHashCursor != "" {
			jsonData["cursor"] = crByPuzzleHashCursor
		}
		if crByPuzzleHashLimit > 0 {
			jsonData["limit"] = crByPuzzleHashLimit
		}
		if crByPuzzleHashOrder != "" {
			jsonData["order"] = crByPuzzleHashOrder
		}
		makeRequest("get_coin_records_by_puzzle_hash", jsonData)
	},
}
