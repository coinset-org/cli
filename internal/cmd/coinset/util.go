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
	return err == nil && r == true
}

func isHex(str string) bool {
	r, err := regexp.MatchString("^(0x)?[0-9A-Fa-f]+$", str)
	return err == nil && r == true
}

func formatHex(str string) string {
	r, _ := regexp.MatchString("^0x[0-9A-Fa-f]+$", str)
	if r == true {
		return str
	}
	return "0x" + str
}

func cleanHex(str string) string {
	if str[:2] == "0x" {
		return str[2:]
	}
	return str
}

func apiRoot() string {
	if api != "" {
		return api
	}
	if testnet == true {
		return "https://testnet10.coinset.org"
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

	byteResult, _ := io.ReadAll(resp.Body)
	processJsonBytes(byteResult)
}

func processJsonData(jsonData map[string]interface{}) {
	query, err := gojq.Parse(jq)
	if err != nil {
		fmt.Println(err)
	}
	iter := query.Run(jsonData) // or query.RunWithContext
	for {
		v, ok := iter.Next()
		if !ok {
			break
		}
		if err, ok := v.(error); ok {
			fmt.Println(err)
		}

		f := colorjson.NewFormatter()
		f.Indent = 2

		s, _ := f.Marshal(v)
		fmt.Println(string(s))
	}
}

func processJsonBytes(jsonBytes []byte) {
	var jsonData map[string]interface{}
	json.Unmarshal(jsonBytes, &jsonData)
	processJsonData(jsonData)
}

func handleRequest(req *http.Request, err error) {

}
