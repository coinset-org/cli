package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	crByHintsIncludeSpentCoins bool
	crByHintsStart             int
	crByHintsEnd               int
)

func init() {
	getCoinRecordsByHintsCmd.Flags().BoolVarP(&crByHintsIncludeSpentCoins, "include-spent-coins", "s", false, "Include spent coins")
	getCoinRecordsByHintsCmd.Flags().IntVarP(&crByHintsStart, "start-height", "", -1, "Start height")
	getCoinRecordsByHintsCmd.Flags().IntVarP(&crByHintsEnd, "end-height", "", -1, "End height")
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
		makeRequest("get_coin_records_by_hints", jsonData)
	},
}
