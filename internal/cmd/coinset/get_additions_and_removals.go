package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	getAdditionsAndRemovalsCursor string
	getAdditionsAndRemovalsLimit  int
	getAdditionsAndRemovalsOrder  string
	getAdditionsAndRemovalsFilter string
)

func init() {
	getAdditionsAndRemovalsCmd.Flags().StringVar(&getAdditionsAndRemovalsCursor, "cursor", "", "Pagination cursor from a previous truncated response")
	getAdditionsAndRemovalsCmd.Flags().IntVar(&getAdditionsAndRemovalsLimit, "limit", 0, "Maximum number of records to return")
	getAdditionsAndRemovalsCmd.Flags().StringVar(&getAdditionsAndRemovalsOrder, "order", "", `Record order: "asc" or "desc"`)
	getAdditionsAndRemovalsCmd.Flags().StringVar(&getAdditionsAndRemovalsFilter, "filter", "", `Filter records: "all", "additions", or "removals"`)
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
		if getAdditionsAndRemovalsLimit < 0 {
			return fmt.Errorf("limit must be >= 1 when provided")
		}
		if getAdditionsAndRemovalsOrder != "" && !isAscOrDesc(getAdditionsAndRemovalsOrder) {
			return fmt.Errorf(`invalid order %q: must be "asc" or "desc"`, getAdditionsAndRemovalsOrder)
		}
		if getAdditionsAndRemovalsFilter != "" &&
			getAdditionsAndRemovalsFilter != "all" &&
			getAdditionsAndRemovalsFilter != "additions" &&
			getAdditionsAndRemovalsFilter != "removals" {
			return fmt.Errorf(`invalid filter %q: must be "all", "additions", or "removals"`, getAdditionsAndRemovalsFilter)
		}
		return nil
	},
	Short: "Retrieves the additions and removals for a certain block",
	Long:  "Retrieves the additions and removals for a certain block by height or header hash. Returns coin records for each addition and removal. Blocks that are not transaction blocks will have empty removal and addition lists. To get the actual puzzles and solutions for spent coins, use the get_puzzle_and_solution api.",
	Run: func(cmd *cobra.Command, args []string) {
		headerHash, _ := convertHeightOrHeaderHash(args[0])
		jsonData := map[string]interface{}{}
		jsonData["header_hash"] = headerHash
		if getAdditionsAndRemovalsCursor != "" {
			jsonData["cursor"] = getAdditionsAndRemovalsCursor
		}
		if getAdditionsAndRemovalsLimit > 0 {
			jsonData["limit"] = getAdditionsAndRemovalsLimit
		}
		if getAdditionsAndRemovalsOrder != "" {
			jsonData["order"] = getAdditionsAndRemovalsOrder
		}
		if getAdditionsAndRemovalsFilter != "" {
			jsonData["filter"] = getAdditionsAndRemovalsFilter
		}
		makeRequest("get_additions_and_removals", jsonData)
	},
}
