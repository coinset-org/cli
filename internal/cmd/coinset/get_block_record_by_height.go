package cmd

import (
    "fmt"
    "strconv"

    "github.com/spf13/cobra"
)

func init() {
    rootCmd.AddCommand(getBlockRecordByHeightCmd)
}


var getBlockRecordByHeightCmd = &cobra.Command{
    Use:   "get_block_record_by_height <height>",
    Args: func(cmd *cobra.Command, args []string) error {
        if err := cobra.ExactArgs(1)(cmd, args); err != nil {
            return err
        }
        if _, err := strconv.Atoi(args[0]); err == nil {
            return nil
        }
        return fmt.Errorf("invalid height specified: %s", args[0])
    },
    Short: "Retrieves a block record by height",
    Long:  `Retrieves a block record by height`,
    Run: func(cmd *cobra.Command, args []string) {
        height, _ := strconv.Atoi(args[0])
        jsonData := map[string]interface{}{}
        jsonData["height"] = height
        makeRequest("get_block_record_by_height", jsonData)
    },
}
