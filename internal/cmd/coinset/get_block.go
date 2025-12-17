package cmd

import (
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(getBlockCmd)
}

var getBlockCmd = &cobra.Command{
	Use: "get_block <height_or_header_hash>",
	Args: func(cmd *cobra.Command, args []string) error {
		if err := cobra.ExactArgs(1)(cmd, args); err != nil {
			return err
		}
		_, err := convertHeightOrHeaderHash(args[0])
		if err != nil {
			return err
		}
		return nil
	},
	Short: "Retrieves an entire block by height or header hash",
	Long:  `Retrieves an entire block by height or header hash`,
	Run: func(cmd *cobra.Command, args []string) {
		headerHash, _ := convertHeightOrHeaderHash(args[0])
		jsonData := map[string]interface{}{}
		jsonData["header_hash"] = headerHash
		makeRequest("get_block", jsonData)
	},
}
