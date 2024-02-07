package cmd

import (
    "fmt"
    "strconv"

    "github.com/spf13/cobra"
)

func init() {
    rootCmd.AddCommand(getBlockRecordsCmd)
}

var getBlockRecordsCmd = &cobra.Command{
    Use:   "get_block_records <start_height> <end_height>",
    Args: func(cmd *cobra.Command, args []string) error {
        if err := cobra.ExactArgs(2)(cmd, args); err != nil {
            return err
        }
        if _, err := strconv.Atoi(args[0]); err != nil {
            return fmt.Errorf("invalid start height specified: %s", args[0])
        }
        if _, err := strconv.Atoi(args[1]); err != nil {
            return fmt.Errorf("invalid end height specified: %s", args[1])
        }
        return nil
    },
    Short: "Gets a list of full blocks by height",
    Long:  `Gets a list of full blocks by height`,
    Run: func(cmd *cobra.Command, args []string) {
        start, _ := strconv.Atoi(args[0])
        end, _ := strconv.Atoi(args[1])
        jsonData := map[string]interface{}{}
        jsonData["start"] = start
        jsonData["end"] = end
        makeRequest("get_block_records", jsonData)
    },
}