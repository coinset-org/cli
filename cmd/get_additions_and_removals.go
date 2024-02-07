package cmd

import (
    "fmt"

    "github.com/spf13/cobra"
)

func init() {
    rootCmd.AddCommand(getAdditionsAndRemovalsCmd)
}

var getAdditionsAndRemovalsCmd = &cobra.Command{
    Use:   "get_additions_and_removals <header_hash>",
    Args: func(cmd *cobra.Command, args []string) error {
        if err := cobra.ExactArgs(1)(cmd, args); err != nil {
            return err
        }
        if isHex(args[0]) == true {
            return nil
        }
        return fmt.Errorf("invalid hex value specified: %s", args[0])
    },
    Short: "Retrieves the additions and removals for a certain block",
    Long:  "Retrieves the additions and removals for a certain block. Returns coin records for each addition and removal. Blocks that are not transaction blocks will have empty removal and addition lists. To get the actual puzzles and solutions for spent coins, use the get_puzzle_and_solution api.",
    Run: func(cmd *cobra.Command, args []string) {
        jsonData := map[string]interface{}{}
        jsonData["header_hash"] = formatHex(args[0])
        makeRequest("get_additions_and_removals", jsonData)
    },
}