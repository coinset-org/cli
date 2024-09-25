package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(getBlockRecordCmd)
}

var getBlockRecordCmd = &cobra.Command{
	Use: "get_block_record <header_hash>",
	Args: func(cmd *cobra.Command, args []string) error {
		if err := cobra.ExactArgs(1)(cmd, args); err != nil {
			return err
		}
		if isHex(args[0]) {
			return nil
		}
		return fmt.Errorf("invalid hex value specified: %s", args[0])
	},
	Short: "Retrieves a block record by header hash",
	Long:  "Retrieves a block record by header hash",
	Run: func(cmd *cobra.Command, args []string) {
		jsonData := map[string]interface{}{}
		jsonData["header_hash"] = formatHex(args[0])
		makeRequest("get_block_record", jsonData)
	},
}
