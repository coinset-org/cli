package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/chia-network/go-chia-libs/pkg/bech32m"
	"github.com/spf13/cobra"
)

func init() {
	addressEncodeCmd.SetHelpFunc(func(command *cobra.Command, strings []string) {
		command.Parent().HelpFunc()(command, strings)
	})

	addressCmd.AddCommand(addressEncodeCmd)
}

var addressEncodeCmd = &cobra.Command{
	Use: "encode <address>",
	Args: func(cmd *cobra.Command, args []string) error {
		if err := cobra.ExactArgs(1)(cmd, args); err != nil {
			return err
		}
		if isAddress(args[0]) == true {
			return nil
		}
		return fmt.Errorf("invalid address value specified: %s", args[0])
	},
	Short: "Encode address to puzzle hash",
	Long:  `Encode address to puzzle hash`,
	Run: func(cmd *cobra.Command, args []string) {
		jsonData := map[string]interface{}{}
		jsonData["address"] = args[0]
		_, puzzleHashBytes, err := bech32m.DecodePuzzleHash(args[0])
		if err != nil {
			fmt.Println(err)
			return
		}
		jsonData["puzzle_hash"] = puzzleHashBytes.String()

		byteResult, _ := json.Marshal(jsonData)
		if raw {
			fmt.Println(string(byteResult))
			return
		}

		processJsonData(jsonData)
	},
}
