package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	crByPuzzleHashIncludeSpentCoins bool
	crByPuzzleHashStart             int
	crByPuzzleHashEnd               int
)

func init() {
	getCoinRecordsByPuzzleHashCmd.Flags().BoolVarP(&crByPuzzleHashIncludeSpentCoins, "include-spent-coins", "s", false, "Include spent coins")
	getCoinRecordsByPuzzleHashCmd.Flags().IntVarP(&crByPuzzleHashStart, "start-height", "", -1, "Start height")
	getCoinRecordsByPuzzleHashCmd.Flags().IntVarP(&crByPuzzleHashEnd, "end-height", "", -1, "End height")
	rootCmd.AddCommand(getCoinRecordsByPuzzleHashCmd)
}

var getCoinRecordsByPuzzleHashCmd = &cobra.Command{
	Use: "get_coin_records_by_puzzle_hash <hash>",
	Args: func(cmd *cobra.Command, args []string) error {
		if err := cobra.ExactArgs(1)(cmd, args); err != nil {
			return err
		}
		if isHex(args[0]) {
			return nil
		}
		return fmt.Errorf("invalid hex value specified: %s", args[0])
	},
	Short: "Retrieves coin records by their puzzle hash",
	Long:  "Retrieves coin records by their puzzle hash",
	Run: func(cmd *cobra.Command, args []string) {
		jsonData := map[string]interface{}{}
		jsonData["puzzle_hash"] = formatHex(args[0])
		if crByPuzzleHashIncludeSpentCoins {
			jsonData["include_spent_coins"] = true
		}
		if crByPuzzleHashStart != -1 {
			jsonData["start_height"] = crByPuzzleHashStart
		}
		if crByPuzzleHashEnd != -1 {
			jsonData["end_height"] = crByPuzzleHashEnd
		}
		makeRequest("get_coin_records_by_puzzle_hash", jsonData)
	},
}
