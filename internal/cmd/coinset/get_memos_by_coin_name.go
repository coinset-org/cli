package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(getMemosByCoinNameCmd)
}

var getMemosByCoinNameCmd = &cobra.Command{
	Use: "get_memos_by_coin_name <name>",
	Args: func(cmd *cobra.Command, args []string) error {
		if err := cobra.ExactArgs(1)(cmd, args); err != nil {
			return err
		}
		if isHex(args[0]) {
			return nil
		}
		return fmt.Errorf("invalid hex value specified: %s", args[0])
	},
	Short: "Retrieves memos for a coin",
	Long:  "Retrieves memos for a coin",
	Run: func(cmd *cobra.Command, args []string) {
		jsonData := map[string]interface{}{}
		jsonData["name"] = formatHex(args[0])
		makeRequest("get_memos_by_coin_name", jsonData)
	},
}
