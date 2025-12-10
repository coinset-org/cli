package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	crByPuzzleHashesIncludeSpentCoins bool
	crByPuzzleHashesStart             int
	crByPuzzleHashesEnd               int
)

func init() {
	getCoinRecordsByPuzzleHashesCmd.Flags().BoolVarP(&crByPuzzleHashesIncludeSpentCoins, "include-spent-coins", "s", false, "Include spent coins")
	getCoinRecordsByPuzzleHashesCmd.Flags().IntVarP(&crByPuzzleHashesStart, "start-height", "", -1, "Start height")
	getCoinRecordsByPuzzleHashesCmd.Flags().IntVarP(&crByPuzzleHashesEnd, "end-height", "", -1, "End height")
	rootCmd.AddCommand(getCoinRecordsByPuzzleHashesCmd)
}

var getCoinRecordsByPuzzleHashesCmd = &cobra.Command{
	Use: "get_coin_records_by_puzzle_hashes <puzzle_hash_or_address> <puzzle_hash_or_address> ...",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return fmt.Errorf("at least one puzzle hash or address is required")
		}
		for _, input := range args {
			_, err := convertAddressOrPuzzleHash(input)
			if err != nil {
				return fmt.Errorf("invalid input '%s': %v", input, err)
			}
		}
		return nil
	},
	Short: "Retrieves coin records by their puzzle hashes or addresses",
	Long:  "Retrieves coin records by their puzzle hashes or addresses",
	Run: func(cmd *cobra.Command, args []string) {
		var puzzleHashes []string
		for _, input := range args {
			puzzleHash, _ := convertAddressOrPuzzleHash(input)
			puzzleHashes = append(puzzleHashes, puzzleHash)
		}
		jsonData := map[string]interface{}{
			"puzzle_hashes": puzzleHashes,
		}
		if crByPuzzleHashesIncludeSpentCoins {
			jsonData["include_spent_coins"] = true
		}
		if crByPuzzleHashesStart != -1 {
			jsonData["start_height"] = crByPuzzleHashesStart
		}
		if crByPuzzleHashesEnd != -1 {
			jsonData["end_height"] = crByPuzzleHashesEnd
		}
		makeRequest("get_coin_records_by_puzzle_hashes", jsonData)
	},
}
