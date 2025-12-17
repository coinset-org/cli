package cmd

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"regexp"
	"strconv"

	"github.com/TylerBrock/colorjson"
	"github.com/chia-network/go-chia-libs/pkg/bech32m"
	"github.com/chia-network/go-chia-libs/pkg/rpc"
	"github.com/chia-network/go-chia-libs/pkg/rpcinterface"
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

func printJson(jsonBytes []byte) {
	query, err := gojq.Parse(jq)
	if err != nil {
		fmt.Println(err)
	}

	var jsonStrings map[string]interface{}
	json.Unmarshal(jsonBytes, &jsonStrings)

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
