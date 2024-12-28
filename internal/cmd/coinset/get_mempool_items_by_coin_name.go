package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	getMempoolItemsByCoinNameIncludeSpentCoins bool
)

func init() {
	getMempoolItemsByCoinNameCmd.Flags().BoolVarP(&getMempoolItemsByCoinNameIncludeSpentCoins, "include-spent-coins", "s", false, "Include items no longer in the mempool")
	rootCmd.AddCommand(getMempoolItemsByCoinNameCmd)
}

var getMempoolItemsByCoinNameCmd = &cobra.Command{
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
		if getMempoolItemsByCoinNameIncludeSpentCoins {
			jsonData["include_spent_coins"] = true
		}
		makeRequest("get_mempool_items_by_coin_name", jsonData)
	},
}
