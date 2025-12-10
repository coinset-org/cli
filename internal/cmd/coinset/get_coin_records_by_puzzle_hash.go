package cmd

import (
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
	Use: "get_coin_records_by_puzzle_hash <puzzle_hash_or_address>",
	Args: func(cmd *cobra.Command, args []string) error {
		if err := cobra.ExactArgs(1)(cmd, args); err != nil {
			return err
		}
		_, err := convertAddressOrPuzzleHash(args[0])
		if err != nil {
			return err
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
		makeRequest("get_coin_records_by_puzzle_hash", jsonData)
	},
}
