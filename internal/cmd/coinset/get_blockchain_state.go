package cmd

import (
    "github.com/spf13/cobra"
)

func init() {
    rootCmd.AddCommand(getBlockchainStateCmd)
}

var getBlockchainStateCmd = &cobra.Command{
    Use:   "get_blockchain_state",
    Short: "Retrieves a summary of the current state of the blockchain and full node",
    Long:  `Retrieves a summary of the current state of the blockchain and full node`,
    Run: func(cmd *cobra.Command, args []string) {
        makeRequest("get_blockchain_state", nil)
    },
}