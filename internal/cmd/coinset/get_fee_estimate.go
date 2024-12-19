package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

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

func init() {
	getFeeEstimateCmd.Flags().StringP("file", "f", "", "Path to file containing the spend bundle JSON")
	getFeeEstimateCmd.Flags().Int64P("cost", "c", 0, "Cost value for fee estimation")
	getFeeEstimateCmd.Flags().String("times", "60,300,600", "Comma-separated list of target times in seconds")
	rootCmd.AddCommand(getFeeEstimateCmd)
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
  coinset get_fee_estimate -c 20000000`,
}

func init() {
	var requestData map[string]interface{}

	getFeeEstimateCmd.Args = func(cmd *cobra.Command, args []string) error {
		dataFile, _ := cmd.Flags().GetString("file")
		cost, _ := cmd.Flags().GetInt64("cost")
		timesStr, _ := cmd.Flags().GetString("times")

		times := []int64{}
		for _, t := range strings.Split(timesStr, ",") {
			time, err := strconv.ParseInt(strings.TrimSpace(t), 10, 64)
			if err != nil {
				return fmt.Errorf("invalid target time: %s", t)
			}
			times = append(times, time)
		}

		requestData = map[string]interface{}{
			"target_times": times,
		}

		if dataFile != "" && cost != 0 {
			return fmt.Errorf("cannot specify both -f and -c flags")
		}
		if dataFile != "" && len(args) > 0 {
			return fmt.Errorf("cannot provide both -f flag and direct argument")
		}
		if cost != 0 && len(args) > 0 {
			return fmt.Errorf("cannot provide both -c flag and direct argument")
		}

		if dataFile != "" {
			data, err := os.ReadFile(dataFile)
			if err != nil {
				return fmt.Errorf("unable to read file %s: %v", dataFile, err)
			}

			var spendBundle map[string]interface{}
			if err := json.Unmarshal(data, &spendBundle); err != nil {
				return fmt.Errorf("invalid JSON in file: %v", err)
			}

			if err := validateSpendBundle(spendBundle); err != nil {
				return fmt.Errorf("invalid spend bundle in file: %v", err)
			}

			requestData["spend_bundle"] = spendBundle
		} else {
			if cost != 0 {
				requestData = map[string]interface{}{
					"cost":         cost,
					"target_times": times,
				}
				return nil
			}

			if len(args) != 1 {
				return fmt.Errorf("must provide either a cost value or spend bundle JSON, or use -f or -c flags")
			}

			if parsedCost, err := strconv.ParseInt(args[0], 10, 64); err == nil {
				requestData = map[string]interface{}{
					"cost":         parsedCost,
					"target_times": times,
				}
			} else {
				var spendBundle map[string]interface{}
				if err := json.Unmarshal([]byte(args[0]), &spendBundle); err == nil {
					if err := validateSpendBundle(spendBundle); err != nil {
						return fmt.Errorf("invalid spend bundle structure: %v", err)
					}
					requestData["spend_bundle"] = spendBundle
				} else {
					if _, err := os.Stat(args[0]); err == nil {
						data, err := os.ReadFile(args[0])
						if err != nil {
							return fmt.Errorf("unable to read file %s: %v", args[0], err)
						}

						if err := json.Unmarshal(data, &spendBundle); err != nil {
							return fmt.Errorf("invalid JSON in file %s: %v", args[0], err)
						}

						if err := validateSpendBundle(spendBundle); err != nil {
							return fmt.Errorf("invalid spend bundle in file %s: %v", args[0], err)
						}

						requestData["spend_bundle"] = spendBundle
					} else {
						return fmt.Errorf("argument must be either a valid number (cost), valid JSON (spend bundle), or a path to a JSON file containing a spend bundle")
					}
				}
			}
		}

		return nil
	}

	getFeeEstimateCmd.Run = func(cmd *cobra.Command, args []string) {
		makeRequest("get_fee_estimate", requestData)
	}
}
