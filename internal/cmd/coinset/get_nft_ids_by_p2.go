package cmd

import "github.com/spf13/cobra"

var (
	getNftIDsByP2IncludeSpentCoins bool
	getNftIDsByP2Cursor            string
	getNftIDsByP2Limit             int
	getNftIDsByP2Order             string
)

func init() {
	getNftIDsByP2Cmd.Flags().BoolVar(&getNftIDsByP2IncludeSpentCoins, "include-spent-coins", false, "Include spent coins")
	getNftIDsByP2Cmd.Flags().StringVar(&getNftIDsByP2Cursor, "cursor", "", "Pagination cursor from a previous response")
	getNftIDsByP2Cmd.Flags().IntVar(&getNftIDsByP2Limit, "limit", 0, "Maximum number of records to return")
	getNftIDsByP2Cmd.Flags().StringVar(&getNftIDsByP2Order, "order", "", `Result order: "asc" or "desc"`)
	rootCmd.AddCommand(getNftIDsByP2Cmd)
}

var getNftIDsByP2Cmd = &cobra.Command{
	Use:   "get_nft_ids_by_p2 <p2_or_address>",
	Short: "List NFT ids by p2",
	Args: func(cmd *cobra.Command, args []string) error {
		if err := cobra.ExactArgs(1)(cmd, args); err != nil {
			return err
		}
		if _, err := normalizeP2Arg(args[0]); err != nil {
			return err
		}
		return applyPaginationOptions(map[string]interface{}{}, getNftIDsByP2Cursor, getNftIDsByP2Limit, getNftIDsByP2Order)
	},
	Run: func(cmd *cobra.Command, args []string) {
		p2, _ := normalizeP2Arg(args[0])
		jsonData := map[string]interface{}{
			"p2": p2,
		}
		if getNftIDsByP2IncludeSpentCoins {
			jsonData["include_spent_coins"] = true
		}
		_ = applyPaginationOptions(jsonData, getNftIDsByP2Cursor, getNftIDsByP2Limit, getNftIDsByP2Order)
		makeRequest("get_nft_ids_by_p2", jsonData)
	},
}
