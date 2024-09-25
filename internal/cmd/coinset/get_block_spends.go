package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(getBlockSpendsCmd)
}

var getBlockSpendsCmd = &cobra.Command{
	Use: "get_block_spends <header_hash>",
	Args: func(cmd *cobra.Command, args []string) error {
		if err := cobra.ExactArgs(1)(cmd, args); err != nil {
			return err
		}
		if isHex(args[0]) {
			return nil
		}
		return fmt.Errorf("invalid hex value specified: %s", args[0])
	},
	Short: "Retrieves every coin that was spent in a block",
	Long:  `Retrieves every coin that was spent in a block`,
	Run: func(cmd *cobra.Command, args []string) {
		jsonData := map[string]interface{}{}
		jsonData["header_hash"] = formatHex(args[0])
		makeRequest("get_block_spends", jsonData)
	},
}
