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
    getCoinRecordsByPuzzleHashesCmd.Flags().IntVarP(&crByPuzzleHashesStart, "start", "", -1, "Start height")
    getCoinRecordsByPuzzleHashesCmd.Flags().IntVarP(&crByPuzzleHashesEnd, "end", "", -1, "End height")
    rootCmd.AddCommand(getCoinRecordsByPuzzleHashesCmd)
}

var getCoinRecordsByPuzzleHashesCmd = &cobra.Command{
    Use:   "get_coin_records_by_puzzle_hashes <hash> <hash> ...",
    Args: func(cmd *cobra.Command, args []string) error {
        if len(args) < 1 {
            return fmt.Errorf("at least one puzzle hash is required")
        }
        for _, name := range args {
            if !isHex(name) {
                return fmt.Errorf("invalid hex value specified: %s", name)
            }
        }
        return nil
    },
    Short: "Retrieves coin records by their puzzle hashes",
    Long:  "Retrieves coin records by their puzzle hashes",
    Run: func(cmd *cobra.Command, args []string) {
        var puzzleHashes []string
        for _, puzzleHash := range args {
            puzzleHashes = append(puzzleHashes, formatHex(puzzleHash))
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
