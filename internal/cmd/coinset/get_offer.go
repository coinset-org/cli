package cmd

import "github.com/spf13/cobra"

func init() {
	rootCmd.AddCommand(getOfferCmd)
}

var getOfferCmd = &cobra.Command{
	Use:   "get_offer <offer_id>",
	Short: "Get offer summary/state by offer id",
	Args: func(cmd *cobra.Command, args []string) error {
		if err := cobra.ExactArgs(1)(cmd, args); err != nil {
			return err
		}
		_, err := normalizeHexArg(args[0], "offer_id")
		return err
	},
	Run: func(cmd *cobra.Command, args []string) {
		offerID, _ := normalizeHexArg(args[0], "offer_id")
		makeRequest("get_offer", map[string]interface{}{
			"offer_id": offerID,
		})
	},
}
