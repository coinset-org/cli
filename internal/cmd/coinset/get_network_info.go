package cmd

import (
    "github.com/spf13/cobra"
)

func init() {
    rootCmd.AddCommand(getNetworkInfoCmd)
}

var getNetworkInfoCmd = &cobra.Command{
    Use:   "get_network_info",
    Short: "Retrieves some information about the current network",
    Long:  `Retrieves some information about the current network`,
    Run: func(cmd *cobra.Command, args []string) {
        makeRequest("get_network_info", nil)
    },
}