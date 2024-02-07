package cmd

import (
    "github.com/spf13/cobra"
)

func init() {
    rootCmd.AddCommand(getAllMempoolItemsCmd)
}

var getAllMempoolItemsCmd = &cobra.Command{
    Use:   "get_all_mempool_items",
    Short: "Returns all items in the mempool",
    Long:  "Returns all items in the mempool",
    Run: func(cmd *cobra.Command, args []string) {
        makeRequest("get_all_mempool_items", nil)
    },
}