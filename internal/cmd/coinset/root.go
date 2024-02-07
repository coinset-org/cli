package cmd

import (
	"os"
    "net/http"

	"github.com/spf13/cobra"
)

var client = &http.Client{}

var rootCmd = &cobra.Command{
	Use:   "coinset",
	Short: "Make Chia RPC requests",
	Long: `Coinset is a hosted Chia API. Use this CLI to make requests to it.`,
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func Get() *cobra.Command {
	return rootCmd
}

func SetVersion(v string) {
	version = v
}

var jq string
var mainnet bool
var testnet bool
var api string
var version = "dev"

func init() {
	rootCmd.PersistentFlags().StringVarP(&jq, "query", "q", ".", "filter to apply using jq syntax")
	rootCmd.PersistentFlags().BoolVar(&mainnet, "mainnet", false, "Use mainnet as the network")
	rootCmd.PersistentFlags().BoolVar(&testnet, "testnet", false, "Use the latest testnet as the network")
	rootCmd.PersistentFlags().StringVarP(&api, "api", "a", "", "api host to use")
	rootCmd.MarkFlagsMutuallyExclusive("mainnet", "testnet", "api")
}
