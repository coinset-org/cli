package cmd

import "github.com/spf13/cobra"

var addressCmd = &cobra.Command{
	Use:   "address",
	Short: "Encode/decode address to/from puzzle hash",
	Long:  `Encode/decode address to/from puzzle hash.`,
}

func init() {
	addressCmd.SetHelpFunc(func(command *cobra.Command, strings []string) {
		command.Flags().MarkHidden("api")
		command.Flags().MarkHidden("mainnet")
		command.Flags().MarkHidden("testnet")
		addressCmd.Parent().HelpFunc()(command, strings)
	})
	rootCmd.AddCommand(addressCmd)
}
