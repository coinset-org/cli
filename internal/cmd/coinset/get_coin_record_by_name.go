package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(getCoinRecordByNameCmd)
}

var getCoinRecordByNameCmd = &cobra.Command{
	Use: "get_coin_record_by_name <name>",
	Args: func(cmd *cobra.Command, args []string) error {
		if err := cobra.ExactArgs(1)(cmd, args); err != nil {
			return err
		}
		if isHex(args[0]) {
			return nil
		}
		return fmt.Errorf("invalid hex value specified: %s", args[0])
	},
	Short: "Retrieves a coin record by its name",
	Long:  "Retrieves a coin record by its name",
	Run: func(cmd *cobra.Command, args []string) {
		jsonData := map[string]interface{}{}
		jsonData["name"] = formatHex(args[0])
		makeRequest("get_coin_record_by_name", jsonData)
	},
}
