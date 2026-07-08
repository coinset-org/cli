package cmd

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
)

var offerStringPattern = regexp.MustCompile(`^offer1[0-9a-z]+$`)

func init() {
	pushOfferCmd.Flags().StringP("file", "f", "", "Path to file containing the offer string")
	rootCmd.AddCommand(pushOfferCmd)
}

var pushOfferCmd = &cobra.Command{
	Use:   "push_offer [offer]",
	Short: "Push an offer (ingest + broadcast)",
	Long: `Push an offer to Coinset for indexing and broadcast. The offer can be provided in three ways:
1. As a direct argument: push_offer offer1...
2. As a file path argument: push_offer ./my_offer.txt
3. Using the -f flag: push_offer -f ./my_offer.txt`,
}

func init() {
	var offer string

	pushOfferCmd.Args = func(cmd *cobra.Command, args []string) error {
		fileFlag, _ := cmd.Flags().GetString("file")

		if fileFlag != "" && len(args) > 0 {
			return fmt.Errorf("cannot provide both -f flag and direct argument")
		}

		if fileFlag == "" && len(args) == 0 {
			return fmt.Errorf("must provide offer either as argument or with -f flag")
		}

		if len(args) > 1 {
			return fmt.Errorf("too many arguments provided")
		}

		var rawOffer string

		if fileFlag != "" {
			data, err := os.ReadFile(fileFlag)
			if err != nil {
				return fmt.Errorf("unable to read file %s: %v", fileFlag, err)
			}
			rawOffer = string(data)
		} else if len(args) == 1 {
			trimmedArg := strings.TrimSpace(args[0])
			if offerStringPattern.MatchString(trimmedArg) {
				rawOffer = trimmedArg
			} else if _, err := os.Stat(args[0]); err == nil {
				data, err := os.ReadFile(args[0])
				if err != nil {
					return fmt.Errorf("unable to read file %s: %v", args[0], err)
				}
				rawOffer = string(data)
			} else {
				rawOffer = args[0]
			}
		}

		offer = strings.TrimSpace(rawOffer)
		if offer == "" {
			return fmt.Errorf("offer must not be empty")
		}
		if !offerStringPattern.MatchString(offer) {
			return fmt.Errorf("invalid offer string: must be a bech32 offer starting with offer1")
		}

		return nil
	}

	pushOfferCmd.Run = func(cmd *cobra.Command, args []string) {
		if inspect {
			log.Fatal("--inspect not supported for push_offer")
		}

		jsonResponse, err := doRpc("push_offer", map[string]interface{}{
			"offer": offer,
		})
		if err != nil {
			log.Fatal(err.Error())
		}
		printJson(jsonResponse)
	}
}
