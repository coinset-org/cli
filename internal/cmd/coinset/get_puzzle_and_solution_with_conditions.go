package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(getPuzzleAndSolutionWithConditionsCmd)
}

var getPuzzleAndSolutionWithConditionsCmd = &cobra.Command{
	Use: "get_puzzle_and_solution_with_conditions <name>",
	Args: func(cmd *cobra.Command, args []string) error {
		if err := cobra.ExactArgs(1)(cmd, args); err != nil {
			return err
		}
		if isHex(args[0]) {
			return nil
		}
		return fmt.Errorf("invalid hex value specified: %s", args[0])
	},
	Short: "Retrieves a coin's spend record by its name including conditions",
	Long:  "Retrieves a coin's spend record by its name including conditions",
	Run: func(cmd *cobra.Command, args []string) {
		jsonData := map[string]interface{}{}
		jsonData["coin_id"] = formatHex(args[0])
		makeRequest("get_puzzle_and_solution_with_conditions", jsonData)
	},
}
