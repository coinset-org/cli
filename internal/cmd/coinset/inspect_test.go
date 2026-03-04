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
