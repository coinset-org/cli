package cmd

import (
	"encoding/json"
	"testing"
)

func TestInspectRpcOutputAddsPuzzleRecognition(t *testing.T) {
	// coin_spends array shape (like block spends); inspectRpcOutput should transform it.
	input := map[string]any{
		"coin_spends": []any{
			map[string]any{
				"coin": map[string]any{
					"parent_coin_info": "0x" + repeat("11", 32),
					"puzzle_hash":      "0x" + repeat("22", 32),
					"amount":           1,
				},
				"puzzle_reveal": "0x01",
				"solution":      "0x80",
			},
		},
	}
	b, _ := json.Marshal(input)
	out, err := inspectRpcOutput(b)
	if err != nil {
		t.Fatalf("inspectRpcOutput failed: %v", err)
	}
	assertHasPuzzleRecognition(t, out)
}

func TestInspectRpcOutputCoinSolutionWrapper(t *testing.T) {
	input := map[string]any{
		"coin_solution": map[string]any{
			"coin": map[string]any{
				"parent_coin_info": "0x" + repeat("11", 32),
				"puzzle_hash":      "0x" + repeat("22", 32),
				"amount":           1,
			},
			"puzzle_reveal": "0x01",
			"solution":      "0x80",
		},
	}
	b, _ := json.Marshal(input)
	out, err := inspectRpcOutput(b)
	if err != nil {
		t.Fatalf("inspectRpcOutput failed: %v", err)
	}
	assertHasPuzzleRecognition(t, out)
}

func TestInspectRpcOutputMempoolItemsArray(t *testing.T) {
	coinSpend := map[string]any{
		"coin": map[string]any{
			"parent_coin_info": "0x" + repeat("11", 32),
			"puzzle_hash":      "0x" + repeat("22", 32),
			"amount":           1,
		},
		"puzzle_reveal": "0x01",
		"solution":      "0x80",
	}
	mempoolItem := map[string]any{
		"spend_bundle": map[string]any{
			"coin_spends":          []any{coinSpend},
			// Valid G2 point encoding for "infinity": 0xc0 followed by 95 zero bytes.
			"aggregated_signature": "0xc0" + repeat("00", 95),
		},
		"fee": 0,
	}
	input := map[string]any{
		"mempool_items": []any{mempoolItem},
	}
	b, _ := json.Marshal(input)
	out, err := inspectRpcOutput(b)
	if err != nil {
		t.Fatalf("inspectRpcOutput failed: %v", err)
	}

	var arr []any
	if err := json.Unmarshal(out, &arr); err != nil {
		t.Fatalf("output not JSON array: %v", err)
	}
	if len(arr) != 1 {
		t.Fatalf("expected 1 inspected item, got %d", len(arr))
	}
	blob, err := json.Marshal(arr[0])
	if err != nil {
		t.Fatalf("failed to re-marshal inspected item: %v", err)
	}
	assertHasPuzzleRecognition(t, blob)
}

func TestInspectRpcOutputBlockSpendsWithConditions(t *testing.T) {
	coinSpend := map[string]any{
		"coin": map[string]any{
			"parent_coin_info": "0x" + repeat("11", 32),
			"puzzle_hash":      "0x" + repeat("22", 32),
			"amount":           1,
		},
		"puzzle_reveal": "0x01",
		"solution":      "0x80",
	}
	input := map[string]any{
		"block_spends_with_conditions": []any{
			map[string]any{
				"coin_spend":  coinSpend,
				"conditions":  []any{},
				"time_received": 0,
			},
		},
	}
	b, _ := json.Marshal(input)
	out, err := inspectRpcOutput(b)
	if err != nil {
		t.Fatalf("inspectRpcOutput failed: %v", err)
	}
	assertHasPuzzleRecognition(t, out)
}

func assertHasPuzzleRecognition(t *testing.T, out []byte) {
	t.Helper()
	var v map[string]any
	if err := json.Unmarshal(out, &v); err != nil {
		t.Fatalf("output not JSON: %v", err)
	}
	result := v["result"].(map[string]any)
	spends := result["spends"].([]any)
	spend0 := spends[0].(map[string]any)
	if _, ok := spend0["puzzle_recognition"]; !ok {
		t.Fatalf("missing puzzle_recognition key; keys=%v", keys(spend0))
	}
}

func repeat(s string, n int) string {
	out := ""
	for i := 0; i < n; i++ {
		out += s
	}
	return out
}

func keys(m map[string]any) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}
