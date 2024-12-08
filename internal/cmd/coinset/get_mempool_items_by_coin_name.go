package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(getMempoolItemsByCoinName)
}

var getMempoolItemsByCoinName = &cobra.Command{
	Use: "get_mempool_items_by_coin_name <coin_name>",
	Args: func(cmd *cobra.Command, args []string) error {
		if err := cobra.ExactArgs(1)(cmd, args); err != nil {
			return err
		}
		if isHex(args[0]) == true {
			return nil
		}
		return fmt.Errorf("invalid hex value specified: %s", args[0])
	},
	Short: "Returns mempool items by coin name",
	Long:  "Returns mempool items by coin name",
	Run: func(cmd *cobra.Command, args []string) {
		jsonData := map[string]interface{}{}
		jsonData["coin_name"] = formatHex(args[0])
		makeRequest("get_mempool_items_by_coin_name", jsonData)
	},
}
