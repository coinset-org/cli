package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	getNftCoinsByNftIDIncludeSpentCoins bool
	getNftCoinsByNftIDCursor            string
	getNftCoinsByNftIDLimit             int
	getNftCoinsByNftIDOrder             string
)

func init() {
	getNftCoinsByNftIDCmd.Flags().BoolVar(&getNftCoinsByNftIDIncludeSpentCoins, "include-spent-coins", false, "Include spent coins")
	getNftCoinsByNftIDCmd.Flags().StringVar(&getNftCoinsByNftIDCursor, "cursor", "", "Pagination cursor from a previous response")
	getNftCoinsByNftIDCmd.Flags().IntVar(&getNftCoinsByNftIDLimit, "limit", 0, "Maximum number of records to return")
	getNftCoinsByNftIDCmd.Flags().StringVar(&getNftCoinsByNftIDOrder, "order", "", `Result order: "asc" or "desc"`)
	rootCmd.AddCommand(getNftCoinsByNftIDCmd)
}

var getNftCoinsByNftIDCmd = &cobra.Command{
	Use:   "get_nft_coins_by_nft_id <nft_id>",
	Short: "List NFT coins by NFT id",
	Args: func(cmd *cobra.Command, args []string) error {
		if err := cobra.ExactArgs(1)(cmd, args); err != nil {
			return err
		}
		if args[0] == "" {
			return fmt.Errorf("nft_id is required")
		}
		return applyPaginationOptions(map[string]interface{}{}, getNftCoinsByNftIDCursor, getNftCoinsByNftIDLimit, getNftCoinsByNftIDOrder)
	},
	Run: func(cmd *cobra.Command, args []string) {
		jsonData := map[string]interface{}{
			"nft_id": normalizeMaybeHex(args[0]),
		}
		if getNftCoinsByNftIDIncludeSpentCoins {
			jsonData["include_spent_coins"] = true
		}
		_ = applyPaginationOptions(jsonData, getNftCoinsByNftIDCursor, getNftCoinsByNftIDLimit, getNftCoinsByNftIDOrder)
		makeRequest("get_nft_coins_by_nft_id", jsonData)
	},
}
