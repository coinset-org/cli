//go:build cgo

package coinsetffi

import (
	"encoding/json"
	"testing"
)

func TestInspectIncludesPuzzleRecognitionKey(t *testing.T) {
	// Minimal spend bundle shape (bytes are toy, recognition may be empty, but key must exist).
	input := map[string]any{
		"spend_bundle": map[string]any{
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
			// Valid G2 point encoding for "infinity": 0xc0 followed by 95 zero bytes.
			"aggregated_signature": "0xc0" + repeat("00", 95),
		},
	}
	b, _ := json.Marshal(input)
	out, err := Inspect(b, false, false)
	if err != nil {
		t.Fatalf("Inspect failed: %v", err)
	}
	var v map[string]any
	if err := json.Unmarshal(out, &v); err != nil {
		t.Fatalf("output not JSON: %v", err)
	}
	result, ok := v["result"].(map[string]any)
	if !ok {
		t.Fatalf("missing result object; out=%s", string(out))
	}
	spendsAny, ok := result["spends"]
	if !ok {
		t.Fatalf("missing spends; out=%s", string(out))
	}
	spends, ok := spendsAny.([]any)
	if !ok || len(spends) == 0 {
		t.Fatalf("spends not array/non-empty; out=%s", string(out))
	}
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
