package cmd

import (
    "fmt"

    "github.com/spf13/cobra"
)

var (
    crByNamesIncludeSpentCoins bool
    crByNamesStart             int
    crByNamesEnd               int
)

func init() {
    getCoinRecordsByNamesCmd.Flags().BoolVarP(&crByNamesIncludeSpentCoins, "include-spent-coins", "s", false, "Include spent coins")
    getCoinRecordsByNamesCmd.Flags().IntVarP(&crByNamesStart, "start", "", -1, "Start height")
    getCoinRecordsByNamesCmd.Flags().IntVarP(&crByNamesEnd, "end", "", -1, "End height")
    rootCmd.AddCommand(getCoinRecordsByNamesCmd)
}

var getCoinRecordsByNamesCmd = &cobra.Command{
    Use:   "get_coin_records_by_names <name1> <name2> ...",
    Args: func(cmd *cobra.Command, args []string) error {
        if len(args) < 1 {
            return fmt.Errorf("at least one name is required")
        }
        for _, name := range args {
            if !isHex(name) {
                return fmt.Errorf("invalid hex value specified: %s", name)
            }
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
        makeRequest("get_coin_records_by_names", jsonData)
    },
}
