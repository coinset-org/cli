package cmd

import (
	"encoding/json"
	"log"
)

func makePrettyRequest(path string, jsonData map[string]interface{}, printPretty func([]byte) error) {
	if jq != "." || raw || describe {
		log.Fatal("--pretty cannot be combined with --query, --raw, or --describe")
	}

	jsonResponse, err := doRpc(path, jsonData)
	if err != nil {
		log.Fatal(err.Error())
	}

	if err := printPretty(jsonResponse); err != nil {
		log.Printf("pretty format unavailable (%v), falling back to JSON output", err)
		printJson(jsonResponse)
	}
}

func responseSucceeded(envelope map[string]json.RawMessage) bool {
	rawSuccess, ok := envelope["success"]
	if !ok {
		return true
	}
	var success bool
	if err := json.Unmarshal(rawSuccess, &success); err != nil {
		return true
	}
	return success
}
