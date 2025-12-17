package cmd

import (
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(getAdditionsAndRemovalsCmd)
}

var getAdditionsAndRemovalsCmd = &cobra.Command{
	Use: "get_additions_and_removals <height_or_header_hash>",
	Args: func(cmd *cobra.Command, args []string) error {
		if err := cobra.ExactArgs(1)(cmd, args); err != nil {
			return err
		}
		_, err := convertHeightOrHeaderHash(args[0])
		if err != nil {
			return err
		}
		return nil
	},
	Short: "Retrieves the additions and removals for a certain block",
	Long:  "Retrieves the additions and removals for a certain block by height or header hash. Returns coin records for each addition and removal. Blocks that are not transaction blocks will have empty removal and addition lists. To get the actual puzzles and solutions for spent coins, use the get_puzzle_and_solution api.",
	Run: func(cmd *cobra.Command, args []string) {
		headerHash, _ := convertHeightOrHeaderHash(args[0])
		jsonData := map[string]interface{}{}
		jsonData["header_hash"] = headerHash
		makeRequest("get_additions_and_removals", jsonData)
	},
}
