package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(getLatestNftCoinByNftIDCmd)
}

var getLatestNftCoinByNftIDCmd = &cobra.Command{
	Use:   "get_latest_nft_coin_by_nft_id <nft_id>",
	Short: "Get latest NFT coin by NFT id",
	Args: func(cmd *cobra.Command, args []string) error {
		if err := cobra.ExactArgs(1)(cmd, args); err != nil {
			return err
		}
		if args[0] == "" {
			return fmt.Errorf("nft_id is required")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		makeRequest("get_latest_nft_coin_by_nft_id", map[string]interface{}{
			"nft_id": normalizeMaybeHex(args[0]),
		})
	},
}
