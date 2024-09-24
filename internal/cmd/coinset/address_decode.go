package cmd

import (
	"fmt"

	"github.com/chia-network/go-chia-libs/pkg/bech32m"
	"github.com/spf13/cobra"
)

func init() {
	addressDecodeCmd.SetHelpFunc(func(command *cobra.Command, strings []string) {
		command.Flags().MarkHidden("testnet")
		command.Parent().HelpFunc()(command, strings)
	})

	addressCmd.AddCommand(addressDecodeCmd)
}

var addressDecodeCmd = &cobra.Command{
	Use: "decode <address>",
	Args: func(cmd *cobra.Command, args []string) error {
		if err := cobra.ExactArgs(1)(cmd, args); err != nil {
			return err
		}
		if isAddress(args[0]) {
			return nil
		}
		return fmt.Errorf("invalid address value specified: %s", args[0])
	},
	Short: "Decode address to puzzle hash",
	Long:  `Decode address to puzzle hash`,
	Run: func(cmd *cobra.Command, args []string) {
		_, puzzleHashBytes, err := bech32m.DecodePuzzleHash(args[0])
		if err != nil {
			fmt.Println(err)
			return
		}

		fmt.Println(puzzleHashBytes.String())
	},
}
