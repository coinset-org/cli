package cmd

import (
    "github.com/spf13/cobra"
)

func init() {
    rootCmd.AddCommand(getBlockCountMetricsCmd)
}

var getBlockCountMetricsCmd = &cobra.Command{
    Use:   "get_block_count_metrics",
    Short: "Gets metrics for the blockchain's blocks",
    Long:  `Gets metrics for the blockchain's blocks`,
    Run: func(cmd *cobra.Command, args []string) {
        makeRequest("get_block_count_metrics", nil)
    },
}