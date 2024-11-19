package cmd

import (
	"fmt"
	"strconv"

	"github.com/chia-network/go-chia-libs/pkg/types"

	"github.com/spf13/cobra"
)

var coin types.Coin

func init() {
	rootCmd.AddCommand(coinIdCmd)
}

var coinIdCmd = &cobra.Command{
	Use: "coin_id <parent_coin_id> <puzzle_hash> <amount>",
	Args: func(cmd *cobra.Command, args []string) error {
		if err := cobra.ExactArgs(3)(cmd, args); err != nil {
			return err
		}

		// Parent
		if !isHex(args[0]) {
			return fmt.Errorf("invalid hex value specified: %s", args[0])
		}
		parent, err := types.Bytes32FromHexString(formatHex(args[0]))
		if err != nil {
			return fmt.Errorf("invalid hex value specified: %s", args[0])
		}
		coin.ParentCoinInfo = parent

		// Puzzle Hash
		if !isHex(args[1]) {
			return fmt.Errorf("invalid hex value specified: %s", args[1])
		}
		puzzle_hash, err := types.Bytes32FromHexString(formatHex(args[1]))
		if err != nil {
			return fmt.Errorf("invalid hex value specified: %s", args[1])
		}
		coin.PuzzleHash = puzzle_hash

		// Amount
		amount, err := strconv.ParseUint(args[2], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid amount: %s", args[2])
		}
		coin.Amount = amount

		return nil
	},
	Short: "Compute a coin id from parent, puzzle and amount",
	Long:  `Compute a coin id from parent, puzzle and amount`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("%s\n", coin.ID().String())
	},
}
