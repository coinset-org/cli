package cmd

import "github.com/spf13/cobra"

var crUseTestnetPrefix bool

var addressCmd = &cobra.Command{
	Use:   "address",
	Short: "Encode/decode address to/from puzzle hash",
	Long:  `Encode/decode address to/from puzzle hash.`,
}

func init() {
	rootCmd.AddCommand(addressCmd)
}
