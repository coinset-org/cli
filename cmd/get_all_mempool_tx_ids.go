package cmd

import (
    "github.com/spf13/cobra"
)

func init() {
    rootCmd.AddCommand(getAllMempoolTxIdsCmd)
}

var getAllMempoolTxIdsCmd = &cobra.Command{
    Use:   "get_all_mempool_tx_ids",
    Short: "Returns all transaction IDs in the mempool",
    Long:  "Returns all transaction IDs in the mempool",
    Run: func(cmd *cobra.Command, args []string) {
        makeRequest("get_all_mempool_tx_ids", nil)
    },
}