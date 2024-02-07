package cmd

import (
    "fmt"

    "github.com/spf13/cobra"
)

func init() {
    rootCmd.AddCommand(getMempoolItemByTxId)
}

var getMempoolItemByTxId = &cobra.Command{
    Use:   "get_mempool_item_by_tx_id <tx_id>",
    Args: func(cmd *cobra.Command, args []string) error {
        if err := cobra.ExactArgs(1)(cmd, args); err != nil {
            return err
        }
        if isHex(args[0]) == true {
            return nil
        }
        return fmt.Errorf("invalid hex value specified: %s", args[0])
    },
    Short: "Returns a mempool item by transaction ID",
    Long:  "Returns a mempool item by transaction ID",
    Run: func(cmd *cobra.Command, args []string) {
        jsonData := map[string]interface{}{}
        jsonData["tx_id"] = formatHex(args[0])
        makeRequest("get_mempool_item_by_tx_id", jsonData)
    },
}