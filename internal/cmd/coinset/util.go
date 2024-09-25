package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"

	"github.com/TylerBrock/colorjson"
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

func apiRoot() string {
	if api != "" {
		return api
	}
	if testnet {
		return "https://testnet11.api.coinset.org"
	}
	return "https://api.coinset.org"
}

func makeRequest(rpc string, jsonData map[string]interface{}) {
	var buf io.Reader
	if jsonData != nil {
		jsonString, _ := json.Marshal(jsonData)
		buf = bytes.NewBuffer([]byte(string(jsonString)))
	}
	req, err := http.NewRequest("POST", apiRoot()+"/"+rpc, buf)
	if err != nil {
		fmt.Println(err)
		return
	}
	req.Header.Add("Content-Type", `application/json"`)

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}

	jsonBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return
	}

	printJson(jsonBytes)
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
