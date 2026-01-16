package cmd

import (
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"net/url"
	"regexp"
	"strconv"
	"sync"
	"time"

	"github.com/TylerBrock/colorjson"
	"github.com/chia-network/go-chia-libs/pkg/bech32m"
	"github.com/chia-network/go-chia-libs/pkg/rpc"
	"github.com/chia-network/go-chia-libs/pkg/rpcinterface"
	"github.com/dustin/go-humanize"
	"github.com/itchyny/gojq"
)

func isAddress(str string) bool {
	r, err := regexp.MatchString("^(xch|txch){1}[0-9A-Za-z]{59}$", str)
	return err == nil && r
}

func isHex(str string) bool {
	r, err := regexp.MatchString("^(0x)?[0-9A-Fa-f]+$", str)
	return err == nil && r
}

func formatHex(str string) string {
	r, _ := regexp.MatchString("^0x[0-9A-Fa-f]+$", str)
	if r {
		return str
	}
	return "0x" + str
}

func convertAddressOrPuzzleHash(input string) (string, error) {
	if isAddress(input) {
		_, puzzleHashBytes, err := bech32m.DecodePuzzleHash(input)
		if err != nil {
			return "", fmt.Errorf("invalid address: %v", err)
		}
		return puzzleHashBytes.String(), nil
	} else if isHex(input) {
		return formatHex(input), nil
	} else {
		return "", fmt.Errorf("invalid input: must be either a Chia address or hex puzzle hash")
	}
}

func convertHeightOrHeaderHash(input string) (string, error) {
	if height, err := strconv.Atoi(input); err == nil {
		return getHeaderHashByHeight(height)
	} else if isHex(input) {
		return formatHex(input), nil
	} else {
		return "", fmt.Errorf("invalid input: must be either a block height (number) or hex header hash")
	}
}

func getHeaderHashByHeight(height int) (string, error) {
	jsonData := map[string]interface{}{
		"height": height,
	}

	jsonResponse, err := doRpc("get_block_record_by_height", jsonData)
	if err != nil {
		return "", fmt.Errorf("failed to get block record by height %d: %v", height, err)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(jsonResponse, &response); err != nil {
		return "", fmt.Errorf("failed to parse response: %v", err)
	}

	blockRecord, ok := response["block_record"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("invalid response format from get_block_record_by_height")
	}

	headerHash, ok := blockRecord["header_hash"].(string)
	if !ok {
		return "", fmt.Errorf("header_hash not found in block record")
	}

	return headerHash, nil
}

func apiHost() string {
	baseUrl, err := url.Parse(apiRoot())
	if err != nil {
		log.Fatal(err.Error())
	}
	return baseUrl.Host
}

func apiRoot() string {
	if api != "" {
		return api
	}
	if testnet {
		return "https://testnet11.api.coinset.org"
	}
	return "https://api.coinset.org"
}

func doRpc(path string, jsonData map[string]interface{}) (json.RawMessage, error) {
	var client *rpc.Client
	var err error

	baseUrl, err := url.Parse(apiRoot())
	if err != nil {
		return nil, fmt.Errorf("failed to parse API URL: %v", err)
	}

	if local {
		client, err = rpc.NewClient(rpc.ConnectionModeHTTP, rpc.WithAutoConfig())
	} else {
		client, err = rpc.NewClient(rpc.ConnectionModePublicHTTP, rpc.WithPublicConfig(), rpc.WithBaseURL(baseUrl))
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create RPC client: %v", err)
	}

	req, err := client.FullNodeService.NewRequest(rpcinterface.Endpoint(path), jsonData)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	jsonResponse := json.RawMessage{}
	_, err = client.FullNodeService.Do(req, &jsonResponse)
	if err != nil {
		return nil, fmt.Errorf("request failed: %v", err)
	}

	return jsonResponse, nil
}

func makeRequest(path string, jsonData map[string]interface{}) {
	jsonResponse, err := doRpc(path, jsonData)
	if err != nil {
		log.Fatal(err.Error())
	}
	printJson(jsonResponse)
}

// Cache for block records (height -> timestamp)
var blockRecordCache = make(map[int]int64)
var blockRecordCacheMutex sync.RWMutex

// Helper function to get current block height (cached)
var cachedBlockHeight *int
var cachedBlockHeightTime time.Time
var cachedBlockHeightMutex sync.RWMutex

func getBlockTimestamp(height int) (int64, error) {
	// Check cache first
	blockRecordCacheMutex.RLock()
	if timestamp, ok := blockRecordCache[height]; ok {
		blockRecordCacheMutex.RUnlock()
		return timestamp, nil
	}
	blockRecordCacheMutex.RUnlock()

	// Fetch block record
	jsonData := map[string]interface{}{
		"height": height,
	}

	jsonResponse, err := doRpc("get_block_record_by_height", jsonData)
	if err != nil {
		return 0, err
	}

	var response map[string]interface{}
	if err := json.Unmarshal(jsonResponse, &response); err != nil {
		return 0, err
	}

	blockRecord, ok := response["block_record"].(map[string]interface{})
	if !ok {
		return 0, fmt.Errorf("invalid response format")
	}

	timestamp, ok := blockRecord["timestamp"].(float64)
	if !ok {
		return 0, fmt.Errorf("timestamp not found")
	}

	timestampInt := int64(timestamp)

	// Cache it
	blockRecordCacheMutex.Lock()
	blockRecordCache[height] = timestampInt
	blockRecordCacheMutex.Unlock()

	return timestampInt, nil
}

func getCurrentBlockHeight() (int, error) {
	cachedBlockHeightMutex.RLock()
	if cachedBlockHeight != nil && time.Since(cachedBlockHeightTime) < 30*time.Second {
		height := *cachedBlockHeight
		cachedBlockHeightMutex.RUnlock()
		return height, nil
	}
	cachedBlockHeightMutex.RUnlock()

	jsonResponse, err := doRpc("get_blockchain_state", nil)
	if err != nil {
		return 0, err
	}

	var response map[string]interface{}
	if err := json.Unmarshal(jsonResponse, &response); err != nil {
		return 0, err
	}

	peak, ok := response["peak"].(map[string]interface{})
	if !ok {
		return 0, fmt.Errorf("invalid response format")
	}

	height, ok := peak["height"].(float64)
	if !ok {
		return 0, fmt.Errorf("height not found")
	}

	heightInt := int(height)
	cachedBlockHeightMutex.Lock()
	cachedBlockHeight = &heightInt
	cachedBlockHeightTime = time.Now()
	cachedBlockHeightMutex.Unlock()
	return heightInt, nil
}

func formatAmount(amount interface{}) string {
	var amountInt int64
	switch v := amount.(type) {
	case float64:
		amountInt = int64(v)
	case int64:
		amountInt = v
	case int:
		amountInt = int64(v)
	default:
		return ""
	}

	// 1 XCH = 1 trillion mojos
	trillion := big.NewInt(1000000000000)
	amountBig := big.NewInt(amountInt)

	// Divide by trillion to get XCH
	xch := new(big.Float).SetInt(amountBig)
	xch.Quo(xch, new(big.Float).SetInt(trillion))

	return fmt.Sprintf("%.12f XCH", xch)
}

func formatTimestampWithRelative(timestamp interface{}) string {
	var ts int64
	switch v := timestamp.(type) {
	case float64:
		ts = int64(v)
	case int64:
		ts = v
	case int:
		ts = int64(v)
	default:
		return ""
	}

	if ts == 0 {
		return "Never"
	}

	blockTime := time.Unix(ts, 0).Local()
	relativeTime := humanize.Time(blockTime)
	absoluteTime := blockTime.Format("2006-01-02 15:04:05")

	return fmt.Sprintf("%s, %s", relativeTime, absoluteTime)
}

func formatBlockHeight(height interface{}) string {
	var heightInt int
	switch v := height.(type) {
	case float64:
		heightInt = int(v)
	case int64:
		heightInt = int(v)
	case int:
		heightInt = v
	default:
		return ""
	}

	// Always fetch the block timestamp to show relative time
	timestamp, err := getBlockTimestamp(heightInt)
	if err != nil {
		// Fallback to estimation if fetch fails
		currentHeight, err := getCurrentBlockHeight()
		if err != nil {
			return fmt.Sprintf("Block %d", heightInt)
		}
		diff := currentHeight - heightInt
		approxMinutes := diff * 18 / 60
		if approxMinutes < 60 {
			return fmt.Sprintf("~%d minutes ago", approxMinutes)
		}
		approxHours := approxMinutes / 60
		if approxHours < 24 {
			return fmt.Sprintf("~%d hours ago", approxHours)
		}
		approxDays := approxHours / 24
		return fmt.Sprintf("~%d days ago", approxDays)
	}

	// Use actual timestamp
	return formatTimestampWithRelative(timestamp)
}

func addDescriptions(data interface{}) interface{} {
	switch v := data.(type) {
	case map[string]interface{}:
		result := make(map[string]interface{})
		for key, value := range v {
			result[key] = addDescriptions(value)

			// Add descriptions for specific fields
			switch key {
			case "amount":
				if desc := formatAmount(value); desc != "" {
					result[key+"_description"] = desc
				}
			case "timestamp":
				if desc := formatTimestampWithRelative(value); desc != "" {
					result[key+"_description"] = desc
				}
			case "confirmed_block_index", "spent_block_index", "block_index", "height":
				if desc := formatBlockHeight(value); desc != "" {
					result[key+"_description"] = desc
				}
			}
		}
		return result
	case []interface{}:
		result := make([]interface{}, len(v))
		for i, item := range v {
			result[i] = addDescriptions(item)
		}
		return result
	default:
		return v
	}
}

func printJson(jsonBytes []byte) {
	query, err := gojq.Parse(jq)
	if err != nil {
		fmt.Println(err)
	}

	var jsonStrings map[string]interface{}
	json.Unmarshal(jsonBytes, &jsonStrings)

	// Add descriptions if flag is set
	if describe {
		jsonStrings = addDescriptions(jsonStrings).(map[string]interface{})
	}

	iter := query.Run(jsonStrings)
	for {
		v, ok := iter.Next()
		if !ok {
			break
		}
		if err, ok := v.(error); ok {
			fmt.Println(err)
		}

		if raw {
			s, _ := json.Marshal(v)
			fmt.Println(string(s))
		} else {
			f := colorjson.NewFormatter()
			f.Indent = 2

			s, _ := f.Marshal(v)
			fmt.Println(string(s))
		}

	}
}
