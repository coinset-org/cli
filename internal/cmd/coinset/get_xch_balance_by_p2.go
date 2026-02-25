package cmd

import "github.com/spf13/cobra"

func init() {
	rootCmd.AddCommand(getXchBalanceByP2Cmd)
}

var getXchBalanceByP2Cmd = &cobra.Command{
	Use:   "get_xch_balance_by_p2 <p2_or_address>",
	Short: "Get XCH balance by p2",
	Args: func(cmd *cobra.Command, args []string) error {
		if err := cobra.ExactArgs(1)(cmd, args); err != nil {
			return err
		}
		_, err := normalizeP2Arg(args[0])
		return err
	},
	Run: func(cmd *cobra.Command, args []string) {
		p2, _ := normalizeP2Arg(args[0])
		makeRequest("get_xch_balance_by_p2", map[string]interface{}{
			"p2": p2,
		})
	},
}
