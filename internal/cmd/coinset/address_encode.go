package cmd

import (
	"encoding/hex"
	"fmt"

	"github.com/chia-network/go-chia-libs/pkg/bech32m"
	"github.com/chia-network/go-chia-libs/pkg/types"
	"github.com/spf13/cobra"
)

func init() {
	addressCmd.AddCommand(addressEncodeCmd)
}

var addressEncodeCmd = &cobra.Command{
	Use: "encode <puzzle_hash>",
	Args: func(cmd *cobra.Command, args []string) error {
		if err := cobra.ExactArgs(1)(cmd, args); err != nil {
			return err
		}
		if isHex(args[0]) {
			return nil
		}
		return fmt.Errorf("invalid hex value specified: %s", args[0])
	},
	Short: "Encode puzzle hash to address",
	Long:  `Encode puzzle hash to address`,
	Run: func(cmd *cobra.Command, args []string) {
		prefix := "xch"
		if testnet {
			prefix = "txch"
		}
		var puzzleHash = formatHex(args[0])
		hexBytes, err := hex.DecodeString(puzzleHash[2:])
		if err != nil {
			fmt.Println(err)
			return
		}

		hexBytes32, err := types.BytesToBytes32(hexBytes)
		if err != nil {
			fmt.Println(err)
			return
		}

		address, err := bech32m.EncodePuzzleHash(hexBytes32, prefix)
		if err != nil {
			fmt.Println(err)
			return
		}

		fmt.Println(address)
	},
}
