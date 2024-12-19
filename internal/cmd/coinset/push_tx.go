package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func init() {
	getPushTxCmd.Flags().StringP("file", "f", "", "Path to file containing the spend bundle JSON")
	rootCmd.AddCommand(getPushTxCmd)
}

var getPushTxCmd = &cobra.Command{
	Use:   "push_tx [spend_bundle_json]",
	Short: "Push spend bundle to the mempool",
	Long: `Push spend bundle to the mempool. The spend bundle can be provided in three ways:
1. As a JSON string argument: push_tx '{"aggregated_signature":"0x...","coin_spends":[...]}'
2. As a file path argument: push_tx ./spend_bundle.json
3. Using the -f flag: push_tx -f ./spend_bundle.json`,
}

func init() {
	var parsedJson map[string]interface{}

	getPushTxCmd.Args = func(cmd *cobra.Command, args []string) error {
		fileFlag, _ := cmd.Flags().GetString("file")

		if fileFlag != "" && len(args) > 0 {
			return fmt.Errorf("cannot provide both -f flag and direct argument")
		}

		if fileFlag == "" && len(args) == 0 {
			return fmt.Errorf("must provide spend bundle either as argument or with -f flag")
		}

		if len(args) > 1 {
			return fmt.Errorf("too many arguments provided. Did you forget to quote your JSON? Expected: push_tx '<json>' or push_tx -f <filename>")
		}

		var spendBundleJson string

		if fileFlag != "" {
			data, err := os.ReadFile(fileFlag)
			if err != nil {
				return fmt.Errorf("unable to read file %s: %v", fileFlag, err)
			}
			spendBundleJson = string(data)
		} else {
			if err := json.Unmarshal([]byte(args[0]), &parsedJson); err == nil {
				if err := validateSpendBundle(parsedJson); err != nil {
					return fmt.Errorf("invalid spend bundle structure: %v", err)
				}
				return nil
			}

			if _, err := os.Stat(args[0]); err == nil {
				data, err := os.ReadFile(args[0])
				if err != nil {
					return fmt.Errorf("unable to read file %s: %v", args[0], err)
				}
				spendBundleJson = string(data)
			} else {
				return fmt.Errorf("argument must be either valid JSON spend bundle or path to a JSON file containing a spend bundle")
			}
		}

		if err := json.Unmarshal([]byte(spendBundleJson), &parsedJson); err != nil {
			return fmt.Errorf("invalid JSON: %v", err)
		}

		if err := validateSpendBundle(parsedJson); err != nil {
			return fmt.Errorf("invalid spend bundle structure: %v", err)
		}

		return nil
	}

	getPushTxCmd.Run = func(cmd *cobra.Command, args []string) {
		requestData := map[string]interface{}{
			"spend_bundle": parsedJson,
		}
		makeRequest("push_tx", requestData)
	}
}
