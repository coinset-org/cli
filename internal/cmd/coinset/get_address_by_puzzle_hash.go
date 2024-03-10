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
	getAddressByPuzzleHashCmd.Flags().BoolVarP(&crUseTestnetPrefix, "use-prefix-txch", "t", false, "use testnet prefix 'txch'")
	getAddressByPuzzleHashCmd.SetHelpFunc(func(command *cobra.Command, strings []string) {
		command.Flags().MarkHidden("api")
		command.Flags().MarkHidden("mainnet")
		command.Flags().MarkHidden("testnet")
		command.Parent().HelpFunc()(command, strings)
	})

	rootCmd.AddCommand(getAddressByPuzzleHashCmd)
}

var getAddressByPuzzleHashCmd = &cobra.Command{
	Use: "get_address_by_puzzle_hash <puzzle_hash>",
	Args: func(cmd *cobra.Command, args []string) error {
		if err := cobra.ExactArgs(1)(cmd, args); err != nil {
			return err
		}
		if isHex(args[0]) == true {
			return nil
		}
		return fmt.Errorf("invalid hex value specified: %s", args[0])
	},
	Short: "Gets an address by puzzle hash",
	Long:  `Gets an address by puzzle hash`,
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
