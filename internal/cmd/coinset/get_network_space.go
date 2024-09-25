package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(getNetworkSpaceCmd)
}

var getNetworkSpaceCmd = &cobra.Command{
	Use: "get_network_space <older_block_header_hash> <newer_block_header_hash>",
	Args: func(cmd *cobra.Command, args []string) error {
		if err := cobra.ExactArgs(2)(cmd, args); err != nil {
			return err
		}
		if !isHex(args[0]) {
			return fmt.Errorf("invalid hex value specified: %s", args[0])
		}
		if !isHex(args[1]) {
			return fmt.Errorf("invalid hex value specified: %s", args[1])
		}
		return nil
	},
	Short: "Retrieves an estimate of the netspace (in bytes)",
	Long:  `Retrieves an estimate of the netspace, which is the total plotted space of all farmers (in bytes)`,
	Run: func(cmd *cobra.Command, args []string) {
		jsonData := map[string]interface{}{}
		jsonData["older_block_header_hash"] = formatHex(args[0])
		jsonData["newer_block_header_hash"] = formatHex(args[1])
		makeRequest("get_network_space", jsonData)
	},
}
