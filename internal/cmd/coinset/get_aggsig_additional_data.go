package cmd

import (
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(getAggsigAdditionalDataCmd)
}

var getAggsigAdditionalDataCmd = &cobra.Command{
	Use:   "get_aggsig_additional_data",
	Short: "Returns the AGG_SIG additional data",
	Long:  `Returns the additional data used for AGG_SIG conditions for the current network`,
	Run: func(cmd *cobra.Command, args []string) {
		makeRequest("get_aggsig_additional_data", nil)
	},
}
