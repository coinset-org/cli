package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func init() {
	getFeeEstimateCmd.Flags().StringP("file", "f", "", "Path to file containing the spend bundle JSON")
	rootCmd.AddCommand(getFeeEstimateCmd)
}

var getFeeEstimateCmd = &cobra.Command{
	Use:   "get_fee_estimate [spend_bundle_json]",
	Short: "Get fee estimate for a spend bundle",
	Long:  `Get fee estimate for a spend bundle. The spend bundle can be provided directly as an argument or via a file using the -f flag.`,
}

func init() {
	var spendBundleJson string
	var parsedJson map[string]interface{}

	getFeeEstimateCmd.Args = func(cmd *cobra.Command, args []string) error {
		fileFlag, _ := cmd.Flags().GetString("file")

		if (len(args) == 0 && fileFlag == "") || (len(args) > 0 && fileFlag != "") {
			return fmt.Errorf("must provide either spend bundle JSON as argument or file path with -f flag, but not both")
		}

		if len(args) > 0 {
			if len(args) > 1 {
				return fmt.Errorf("too many arguments provided. Did you forget to quote your JSON? Expected: get_fee_estimate '<json>' or get_fee_estimate -f <filename>")
			}
			spendBundleJson = args[0]
		}

		if fileFlag != "" {
			data, err := os.ReadFile(fileFlag)
			if err != nil {
				return fmt.Errorf("unable to read file %s: %v", fileFlag, err)
			}
			spendBundleJson = string(data)
		}

		if err := json.Unmarshal([]byte(spendBundleJson), &parsedJson); err != nil {
			return fmt.Errorf("invalid JSON: %v", err)
		}

		return nil
	}

	getFeeEstimateCmd.Run = func(cmd *cobra.Command, args []string) {
		makeRequest("get_fee_estimate", parsedJson)
	}
}
