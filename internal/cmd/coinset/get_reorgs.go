package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	getReorgsLimit  int
	getReorgsCursor string
	getReorgsPretty bool
)

func init() {
	getReorgsCmd.Flags().IntVar(&getReorgsLimit, "limit", 0, "Maximum number of reorg events to return (1-200)")
	getReorgsCmd.Flags().StringVar(&getReorgsCursor, "cursor", "", "Pagination cursor from a previous truncated response")
	getReorgsCmd.Flags().BoolVarP(&getReorgsPretty, "pretty", "p", false, "display formatted reorg output")
	rootCmd.AddCommand(getReorgsCmd)
}

var getReorgsCmd = &cobra.Command{
	Use:   "get_reorgs",
	Short: "List recent chain reorg events",
	Long:  "Returns persisted chain reorg events in reverse chronological order.",
	Args: func(cmd *cobra.Command, args []string) error {
		if err := cobra.NoArgs(cmd, args); err != nil {
			return err
		}
		if getReorgsLimit != 0 && (getReorgsLimit < 1 || getReorgsLimit > 200) {
			return fmt.Errorf("limit must be between 1 and 200")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		jsonData := map[string]interface{}{}

		if getReorgsLimit != 0 {
			jsonData["limit"] = getReorgsLimit
		}
		if getReorgsCursor != "" {
			jsonData["cursor"] = getReorgsCursor
		}

		makeReorgRequest("get_reorgs", jsonData, getReorgsPretty)
	},
}
