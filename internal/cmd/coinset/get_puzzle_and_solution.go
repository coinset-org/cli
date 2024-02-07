package cmd

import (
    "fmt"

    "github.com/spf13/cobra"
)

func init() {
    rootCmd.AddCommand(getPuzzleAndSolutionCmd)
}

var getPuzzleAndSolutionCmd = &cobra.Command{
    Use:   "get_puzzle_and_solution <name>",
    Args: func(cmd *cobra.Command, args []string) error {
        if err := cobra.ExactArgs(1)(cmd, args); err != nil {
            return err
        }
        if isHex(args[0]) == true {
            return nil
        }
        return fmt.Errorf("invalid hex value specified: %s", args[0])
    },
    Short: "Retrieves a coin's spend record by its name",
    Long:  "Retrieves a coin's spend record by its name",
    Run: func(cmd *cobra.Command, args []string) {
        jsonData := map[string]interface{}{}
        jsonData["coin_id"] = formatHex(args[0])
        makeRequest("get_puzzle_and_solution", jsonData)
    },
}