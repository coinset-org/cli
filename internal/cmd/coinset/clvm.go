package cmd

import (
	"fmt"
	"log"

	"github.com/coinset-org/cli/internal/coinsetffi"
	"github.com/spf13/cobra"
)

var clvmCmd = &cobra.Command{
	Use:   "clvm",
	Short: "CLVM utilities (compile/decompile/run)",
}

func init() {
	rootCmd.AddCommand(clvmCmd)

	clvmCmd.AddCommand(clvmDecompileCmd)
	clvmCmd.AddCommand(clvmCompileCmd)
	clvmCmd.AddCommand(clvmRunCmd)

	clvmRunCmd.Flags().String("program", "", "CLVM program (hex bytes or s-expression)")
	clvmRunCmd.Flags().String("env", "()", "CLVM environment (hex bytes or s-expression)")
	clvmRunCmd.Flags().Uint64("max-cost", 0, "Maximum cost (0 = unlimited)")
	clvmRunCmd.Flags().Bool("cost", false, "Include cost in output")
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
