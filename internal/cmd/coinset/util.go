package cmd

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"regexp"

	"github.com/TylerBrock/colorjson"
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

func apiHost() string {
	if testnet {
		return "testnet11.api.coinset.org"
	}
	return "api.coinset.org"
}

func apiRoot() string {
	if api != "" {
		return api
	}
	return fmt.Sprintf("https://%s", apiHost())
}

func makeRequest(path string, jsonData map[string]interface{}) {
	var client *rpc.Client
	var err error

	if local {
		client, err = rpc.NewClient(rpc.ConnectionModeHTTP, rpc.WithAutoConfig())
	} else {
		client, err = rpc.NewClient(rpc.ConnectionModePublicHTTP, rpc.WithPublicConfig(), rpc.WithBaseURL(&url.URL{
			Scheme: "https",
			Host:   apiHost(),
		}))
	}

	if err != nil {
		log.Fatal(err.Error())
	}

	req, err := client.FullNodeService.NewRequest(rpcinterface.Endpoint(path), jsonData)
	if err != nil {
		log.Fatal(err.Error())
	}

	jsonResponse := json.RawMessage{}
	_, err = client.FullNodeService.Do(req, &jsonResponse)
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
