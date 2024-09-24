package cmd

import (
	"fmt"

	"github.com/chia-network/go-chia-libs/pkg/bech32m"
	"github.com/spf13/cobra"
)

func init() {
	addressDecodeCmd.SetHelpFunc(func(command *cobra.Command, strings []string) {
		command.Flags().MarkHidden("api")
		command.Flags().MarkHidden("mainnet")
		command.Flags().MarkHidden("testnet")
		command.Flags().MarkHidden("raw")
		command.Flags().MarkHidden("query")
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
	Short: "Decode puzzle hash to address",
	Long:  `Decode puzzle hash to address`,
	Run: func(cmd *cobra.Command, args []string) {
		_, puzzleHashBytes, err := bech32m.DecodePuzzleHash(args[0])
		if err != nil {
			fmt.Println(err)
			return
		}

		fmt.Println(puzzleHashBytes.String())
	},
}
