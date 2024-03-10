package cmd

import (
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/chia-network/go-chia-libs/pkg/bech32m"
	"github.com/chia-network/go-chia-libs/pkg/types"
	"github.com/spf13/cobra"
)

var crUseTestnetPrefix bool

func init() {
	addressDecodeCmd.Flags().BoolVarP(&crUseTestnetPrefix, "use-prefix-txch", "t", false, "use testnet prefix 'txch'")
	addressDecodeCmd.SetHelpFunc(func(command *cobra.Command, strings []string) {
		command.Parent().HelpFunc()(command, strings)
	})

	addressCmd.AddCommand(addressDecodeCmd)
}

var addressDecodeCmd = &cobra.Command{
	Use: "decode <puzzle_hash>",
	Args: func(cmd *cobra.Command, args []string) error {
		if err := cobra.ExactArgs(1)(cmd, args); err != nil {
			return err
		}
		if isHex(args[0]) == true {
			return nil
		}
		return fmt.Errorf("invalid hex value specified: %s", args[0])
	},
	Short: "Decode address from puzzle hash",
	Long:  `Decode address from puzzle hash`,
	Run: func(cmd *cobra.Command, args []string) {
		prefix := "xch"
		if crUseTestnetPrefix {
			prefix = "txch"
		}
		jsonData := map[string]interface{}{}
		jsonData["puzzle_hash"] = formatHex(args[0])

		hexBytes, err := hex.DecodeString(cleanHex(args[0]))
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
		jsonData["address"] = address

		byteResult, _ := json.Marshal(jsonData)
		if raw {
			fmt.Println(string(byteResult))
			return
		}

		processJsonData(jsonData)
	},
}
