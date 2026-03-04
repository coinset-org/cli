package cmd

import (
	"encoding/hex"
	"fmt"
	"log"
	"strings"

	"github.com/coinset-org/cli/internal/coinsetffi"
	"github.com/spf13/cobra"
)

var clvmCmd = &cobra.Command{
	Use:   "clvm",
	Short: "CLVM utilities (compile/decompile/run/curry/uncurry/tree_hash)",
}

func init() {
	rootCmd.AddCommand(clvmCmd)

	clvmCmd.AddCommand(clvmDecompileCmd)
	clvmCmd.AddCommand(clvmCompileCmd)
	clvmCmd.AddCommand(clvmRunCmd)
	clvmCmd.AddCommand(clvmTreeHashCmd)
	clvmCmd.AddCommand(clvmCurryCmd)
	clvmCmd.AddCommand(clvmUncurryCmd)

	clvmRunCmd.Flags().String("program", "", "CLVM program (hex bytes or s-expression)")
	clvmRunCmd.Flags().String("env", "()", "CLVM environment (hex bytes or s-expression)")
	clvmRunCmd.Flags().Uint64("max-cost", 0, "Maximum cost (0 = unlimited)")
	clvmRunCmd.Flags().Bool("cost", false, "Include cost in output")

	clvmCurryCmd.Flags().StringArray("atom", nil, "Curry arg as raw atom bytes (repeatable)")
	clvmCurryCmd.Flags().StringArray("tree-hash", nil, "Curry arg as pre-computed tree hash (repeatable)")
	clvmCurryCmd.Flags().StringArray("program", nil, "Curry arg as serialized CLVM program (repeatable)")
}

var clvmDecompileCmd = &cobra.Command{
	Use:   "decompile <hex_bytes>",
	Short: "Decode CLVM bytes to readable CLVM",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		out, err := coinsetffi.ClvmDecompile(args[0], false)
		if err != nil {
			log.Fatal(err.Error())
		}
		printJson(out)
	},
}

var clvmCompileCmd = &cobra.Command{
	Use:   "compile <clvm>",
	Short: "Encode readable CLVM to bytes",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		out, err := coinsetffi.ClvmCompile(args[0], false)
		if err != nil {
			log.Fatal(err.Error())
		}
		printJson(out)
	},
}

var clvmRunCmd = &cobra.Command{
	Use:   "run [program] [env]",
	Short: "Run CLVM program with environment",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) > 2 {
			return fmt.Errorf("too many arguments")
		}
		if len(args) == 0 {
			prog, _ := cmd.Flags().GetString("program")
			if prog == "" {
				return fmt.Errorf("provide program as arg or via --program")
			}
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		progFlag, _ := cmd.Flags().GetString("program")
		envFlag, _ := cmd.Flags().GetString("env")
		maxCost, _ := cmd.Flags().GetUint64("max-cost")
		includeCost, _ := cmd.Flags().GetBool("cost")

		program := progFlag
		env := envFlag
		if len(args) >= 1 {
			program = args[0]
		}
		if len(args) >= 2 {
			env = args[1]
		}

		out, err := coinsetffi.ClvmRun(program, env, maxCost, includeCost, false)
		if err != nil {
			log.Fatal(err.Error())
		}
		printJson(out)
	},
}

var clvmTreeHashCmd = &cobra.Command{
	Use:   "tree_hash <program>",
	Short: "Compute CLVM tree hash (hex bytes or s-expression)",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		out, err := coinsetffi.ClvmTreeHash(args[0], false)
		if err != nil {
			log.Fatal(err.Error())
		}
		printJson(out)
	},
}

var clvmCurryCmd = &cobra.Command{
	Use:   "curry <mod> [arg1] [arg2] ... [--atom val] [--tree-hash val] [--program val]",
	Short: "Curry arguments into a CLVM program (inputs are hex bytes or s-expressions)",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		modInput := args[0]
		modIsHash := isThirtyTwoByteHex(modInput)

		var curryArgs []coinsetffi.CurryArg

		for _, a := range args[1:] {
			kind := "program"
			if isThirtyTwoByteHex(a) {
				kind = "atom"
			}
			curryArgs = append(curryArgs, coinsetffi.CurryArg{Kind: kind, Value: a})
		}

		atomArgs, _ := cmd.Flags().GetStringArray("atom")
		for _, v := range atomArgs {
			curryArgs = append(curryArgs, coinsetffi.CurryArg{Kind: "atom", Value: v})
		}

		treeHashArgs, _ := cmd.Flags().GetStringArray("tree-hash")
		for _, v := range treeHashArgs {
			curryArgs = append(curryArgs, coinsetffi.CurryArg{Kind: "tree_hash", Value: v})
		}

		programArgs, _ := cmd.Flags().GetStringArray("program")
		for _, v := range programArgs {
			curryArgs = append(curryArgs, coinsetffi.CurryArg{Kind: "program", Value: v})
		}

		out, err := coinsetffi.ClvmCurry(modInput, modIsHash, curryArgs, false)
		if err != nil {
			log.Fatal(err.Error())
		}
		printJson(out)
	},
}

func isThirtyTwoByteHex(s string) bool {
	raw := strings.TrimPrefix(strings.TrimPrefix(s, "0x"), "0X")
	if len(raw) != 64 {
		return false
	}
	_, err := hex.DecodeString(raw)
	return err == nil
}

var clvmUncurryCmd = &cobra.Command{
	Use:   "uncurry <program>",
	Short: "Uncurry a CLVM program (hex bytes or s-expression)",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		out, err := coinsetffi.ClvmUncurry(args[0], false)
		if err != nil {
			log.Fatal(err.Error())
		}
		printJson(out)
	},
}
