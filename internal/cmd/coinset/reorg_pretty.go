package cmd

import (
	"encoding/json"
	"fmt"
	"strings"
)

type reorgEvent struct {
	ID            string `json:"id"`
	DetectedAtMS  int64  `json:"detected_at_ms"`
	OldPeakHeight int64  `json:"old_peak_height"`
	OldPeakHash   string `json:"old_peak_hash"`
	NewPeakHeight int64  `json:"new_peak_height"`
	NewPeakHash   string `json:"new_peak_hash"`
	ReorgDepth    int64  `json:"reorg_depth"`
}

type getReorgsEnvelope struct {
	Reorgs     []reorgEvent `json:"reorgs"`
	Truncated  *bool        `json:"truncated"`
	NextCursor string       `json:"next_cursor"`
	Success    bool         `json:"success"`
}

func makeReorgRequest(path string, jsonData map[string]interface{}, pretty bool) {
	if !pretty {
		makeRequest(path, jsonData)
		return
	}

	makePrettyRequest(path, jsonData, printPrettyReorgResponse)
}

func printPrettyReorgResponse(jsonBytes []byte) error {
	var envelope map[string]json.RawMessage
	if err := json.Unmarshal(jsonBytes, &envelope); err != nil {
		return err
	}

	if !responseSucceeded(envelope) {
		printJson(jsonBytes)
		return nil
	}

	var payload getReorgsEnvelope
	if err := json.Unmarshal(jsonBytes, &payload); err != nil {
		return err
	}

	printReorgList(payload, newANSIStyler())
	return nil
}

func printReorgList(payload getReorgsEnvelope, styler ansiStyler) {
	if len(payload.Reorgs) == 0 {
		fmt.Println("No reorgs.")
		return
	}

	for idx, reorg := range payload.Reorgs {
		if idx > 0 {
			fmt.Println()
		}
		fmt.Println(formatReorgRecord(reorg, styler))
	}

	if payload.Truncated != nil && *payload.Truncated {
		fmt.Printf("\nMore results available. Next cursor: %s\n", payload.NextCursor)
	}
}

func formatReorgRecord(reorg reorgEvent, styler ansiStyler) string {
	var out strings.Builder

	oldHeight := formatSignedStringWithCommas(fmt.Sprintf("%d", reorg.OldPeakHeight))
	newHeight := formatSignedStringWithCommas(fmt.Sprintf("%d", reorg.NewPeakHeight))
	depth := formatSignedStringWithCommas(fmt.Sprintf("%d", reorg.ReorgDepth))
	age := formatAge(reorg.DetectedAtMS)

	fmt.Fprintf(
		&out,
		"%s  %s %s   %s -> %s   %s %s\n",
		styler.bold("REORG"),
		styler.dim("depth"),
		depth,
		oldHeight,
		newHeight,
		styler.dim("age"),
		age,
	)

	out.WriteString(formatSection("Peaks", []string{
		"old  " + shortenHexIdentifier(reorg.OldPeakHash, 10, 8),
		"new  " + shortenHexIdentifier(reorg.NewPeakHash, 10, 8),
	}, styler))

	return strings.TrimRight(out.String(), "\n")
}
