package cmd

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(getFeeTiersByCostCmd)
}

var getFeeTiersByCostCmd = &cobra.Command{
	Use:   "get_fee_tiers_by_cost <cost>",
	Short: "Get wallet-friendly fee tiers for a cost",
	Long:  "Get economy, standard, and urgent fee tiers for a transaction cost",
	Args: func(cmd *cobra.Command, args []string) error {
		if err := cobra.ExactArgs(1)(cmd, args); err != nil {
			return err
		}
		cost, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid cost specified: %s", args[0])
		}
		if cost < 0 {
			return fmt.Errorf("cost must be >= 0")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		cost, _ := strconv.ParseInt(args[0], 10, 64)
		jsonData := map[string]interface{}{
			"cost": cost,
		}
		makeRequest("get_fee_tiers_by_cost", jsonData)
	},
}
