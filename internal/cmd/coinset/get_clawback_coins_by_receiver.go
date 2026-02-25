package cmd

import "github.com/spf13/cobra"

var (
	getClawbackCoinsByReceiverCursor string
	getClawbackCoinsByReceiverLimit  int
	getClawbackCoinsByReceiverOrder  string
)

func init() {
	getClawbackCoinsByReceiverCmd.Flags().StringVar(&getClawbackCoinsByReceiverCursor, "cursor", "", "Pagination cursor from a previous response")
	getClawbackCoinsByReceiverCmd.Flags().IntVar(&getClawbackCoinsByReceiverLimit, "limit", 0, "Maximum number of records to return")
	getClawbackCoinsByReceiverCmd.Flags().StringVar(&getClawbackCoinsByReceiverOrder, "order", "", `Result order: "asc" or "desc"`)
	rootCmd.AddCommand(getClawbackCoinsByReceiverCmd)
}

var getClawbackCoinsByReceiverCmd = &cobra.Command{
	Use:   "get_clawback_coins_by_receiver <p2_or_address>",
	Short: "List clawback coins by receiver p2",
	Args: func(cmd *cobra.Command, args []string) error {
		if err := cobra.ExactArgs(1)(cmd, args); err != nil {
			return err
		}
		if _, err := normalizeP2Arg(args[0]); err != nil {
			return err
		}
		return applyPaginationOptions(map[string]interface{}{}, getClawbackCoinsByReceiverCursor, getClawbackCoinsByReceiverLimit, getClawbackCoinsByReceiverOrder)
	},
	Run: func(cmd *cobra.Command, args []string) {
		p2, _ := normalizeP2Arg(args[0])
		jsonData := map[string]interface{}{
			"p2": p2,
		}
		_ = applyPaginationOptions(jsonData, getClawbackCoinsByReceiverCursor, getClawbackCoinsByReceiverLimit, getClawbackCoinsByReceiverOrder)
		makeRequest("get_clawback_coins_by_receiver", jsonData)
	},
}
