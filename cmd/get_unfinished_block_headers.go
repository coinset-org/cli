package cmd

import (
    "github.com/spf13/cobra"
)

func init() {
    rootCmd.AddCommand(getUnfinishedBlockHeadersCmd)
}

var getUnfinishedBlockHeadersCmd = &cobra.Command{
    Use:   "get_unfinished_block_headers",
    Short: "Retrieves recent unfinished header blocks",
    Long:  `Retrieves recent unfinished header blocks`,
    Run: func(cmd *cobra.Command, args []string) {
        makeRequest("get_unfinished_block_headers", nil)
    },
}