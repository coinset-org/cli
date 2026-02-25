package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

var supportedSpendTypes = map[string]struct{}{
	"send_xch_transaction":   {},
	"cat_spend":              {},
	"take_offer":             {},
	"cancel_offer":           {},
	"nft_set_nft_did":        {},
	"nft_transfer_nft":       {},
	"create_new_pool_wallet": {},
	"pw_absorb_rewards":      {},
	"create_new_did_wallet":  {},
}

func validateSpendBundle(data map[string]interface{}) error {
	aggregatedSig, hasAggSig := data["aggregated_signature"].(string)
	if !hasAggSig {
		return fmt.Errorf("spend bundle missing or invalid aggregated_signature field")
	}
	if !strings.HasPrefix(aggregatedSig, "0x") {
		return fmt.Errorf("aggregated_signature must start with 0x")
	}

	coinSpends, hasCoinSpends := data["coin_spends"].([]interface{})
	if !hasCoinSpends {
		return fmt.Errorf("spend bundle missing or invalid coin_spends field")
	}

	for i, spend := range coinSpends {
		spendMap, ok := spend.(map[string]interface{})
		if !ok {
			return fmt.Errorf("invalid coin spend at index %d", i)
		}

		coin, hasCoin := spendMap["coin"].(map[string]interface{})
		if !hasCoin {
			return fmt.Errorf("missing or invalid coin field in coin spend at index %d", i)
		}

		required := []string{"amount", "parent_coin_info", "puzzle_hash"}
		for _, field := range required {
			val, exists := coin[field]
			if !exists {
				return fmt.Errorf("coin missing required field %s at index %d", field, i)
			}
			if field != "amount" {
				// Validate hex fields
				hexStr, ok := val.(string)
				if !ok || !strings.HasPrefix(hexStr, "0x") {
					return fmt.Errorf("coin field %s must be a hex string starting with 0x at index %d", field, i)
				}
			}
		}

		required = []string{"puzzle_reveal", "solution"}
		for _, field := range required {
			val, exists := spendMap[field].(string)
			if !exists {
				return fmt.Errorf("coin spend missing required field %s at index %d", field, i)
			}
			if !strings.HasPrefix(val, "0x") {
				return fmt.Errorf("%s must start with 0x at index %d", field, i)
			}
		}
	}

	return nil
}

func isValidSpendType(spendType string) bool {
	_, ok := supportedSpendTypes[spendType]
	return ok
}

func parseTargetTimes(timesStr string) ([]int64, error) {
	times := []int64{}
	for _, t := range strings.Split(timesStr, ",") {
		value, err := strconv.ParseInt(strings.TrimSpace(t), 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid target time: %s", t)
		}
		times = append(times, value)
	}
	return times, nil
}

func loadSpendBundleFromFile(path string) (map[string]interface{}, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("unable to read file %s: %v", path, err)
	}

	var spendBundle map[string]interface{}
	if err := json.Unmarshal(data, &spendBundle); err != nil {
		return nil, fmt.Errorf("invalid JSON in file %s: %v", path, err)
	}

	if err := validateSpendBundle(spendBundle); err != nil {
		return nil, fmt.Errorf("invalid spend bundle in file %s: %v", path, err)
	}

	return spendBundle, nil
}

var getFeeEstimateCmd = &cobra.Command{
	Use:   "get_fee_estimate [cost_or_spend_bundle]",
	Short: "Get fee estimate based on cost or spend bundle",
	Long: `Get fee estimate based on either a cost value or spend bundle.
Examples:
  coinset get_fee_estimate 20000000
  coinset get_fee_estimate spend_bundle.json
  coinset get_fee_estimate '{"coin_spends":[...]}'
  coinset get_fee_estimate -f spend_bundle.json
  coinset get_fee_estimate -f spend_bundle.json --times 60,120,300
  coinset get_fee_estimate -c 20000000
  coinset get_fee_estimate --spend-type send_xch_transaction --times 60,300
  coinset get_fee_estimate --spend-type take_offer --spend-count 2`,
}

func init() {
	var requestData map[string]interface{}

	getFeeEstimateCmd.Flags().StringP("file", "f", "", "Path to file containing the spend bundle JSON")
	getFeeEstimateCmd.Flags().Int64P("cost", "c", 0, "Cost value for fee estimation")
	getFeeEstimateCmd.Flags().String("times", "60,300,600", "Comma-separated list of target times in seconds")
	getFeeEstimateCmd.Flags().String("spend-type", "", "Spend type for fee estimation")
	getFeeEstimateCmd.Flags().Int64("spend-count", 1, "Number of spends represented by the request")
	rootCmd.AddCommand(getFeeEstimateCmd)

	getFeeEstimateCmd.Args = func(cmd *cobra.Command, args []string) error {
		dataFile, _ := cmd.Flags().GetString("file")
		cost, _ := cmd.Flags().GetInt64("cost")
		timesStr, _ := cmd.Flags().GetString("times")
		spendType, _ := cmd.Flags().GetString("spend-type")
		spendCount, _ := cmd.Flags().GetInt64("spend-count")

		if spendCount < 1 {
			return fmt.Errorf("spend-count must be >= 1")
		}

		times, err := parseTargetTimes(timesStr)
		if err != nil {
			return err
		}
		requestData = map[string]interface{}{
			"target_times": times,
		}
		if spendCount != 1 {
			requestData["spend_count"] = spendCount
		}

		if cost < 0 {
			return fmt.Errorf("cost must be >= 0")
		}

		specifiedModes := 0
		if dataFile != "" {
			specifiedModes++
		}
		if cost != 0 {
			specifiedModes++
		}
		if spendType != "" {
			specifiedModes++
		}
		if specifiedModes > 1 {
			return fmt.Errorf("choose exactly one of --file, --cost, or --spend-type")
		}

		if len(args) > 1 {
			return fmt.Errorf("too many arguments provided")
		}

		if spendType != "" {
			if len(args) > 0 {
				return fmt.Errorf("cannot provide positional argument with --spend-type")
			}
			if !isValidSpendType(spendType) {
				return fmt.Errorf("invalid spend-type: %s", spendType)
			}
			requestData["spend_type"] = spendType
			return nil
		}

		if dataFile != "" {
			if len(args) > 0 {
				return fmt.Errorf("cannot provide both --file and positional argument")
			}
			spendBundle, err := loadSpendBundleFromFile(dataFile)
			if err != nil {
				return err
			}
			requestData["spend_bundle"] = spendBundle
			return nil
		}

		if cost != 0 {
			requestData["cost"] = cost
			return nil
		}

		if len(args) != 1 {
			return fmt.Errorf("must provide one of --file, --cost, --spend-type, or a positional cost/spend bundle")
		}

		if parsedCost, err := strconv.ParseInt(args[0], 10, 64); err == nil {
			if parsedCost < 0 {
				return fmt.Errorf("cost must be >= 0")
			}
			requestData["cost"] = parsedCost
			return nil
		}

		var spendBundle map[string]interface{}
		if err := json.Unmarshal([]byte(args[0]), &spendBundle); err == nil {
			if err := validateSpendBundle(spendBundle); err != nil {
				return fmt.Errorf("invalid spend bundle structure: %v", err)
			}
			requestData["spend_bundle"] = spendBundle
			return nil
		}

		if _, err := os.Stat(args[0]); err == nil {
			spendBundle, err := loadSpendBundleFromFile(args[0])
			if err != nil {
				return err
			}
			requestData["spend_bundle"] = spendBundle
			return nil
		}

		return fmt.Errorf("argument must be either a valid number (cost), valid JSON (spend bundle), or a path to a JSON file containing a spend bundle")
	}

	getFeeEstimateCmd.Run = func(cmd *cobra.Command, args []string) {
		makeRequest("get_fee_estimate", requestData)
	}
}
