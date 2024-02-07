package cmd

import (
	"bytes"
    "encoding/json"
    "fmt"
    "io"
	"regexp"
    "io/ioutil"
    "net/http"

	"github.com/itchyny/gojq"
	"github.com/TylerBrock/colorjson"
)

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

    var result map[string]interface{}
    byteResult, _ := ioutil.ReadAll(resp.Body)
    json.Unmarshal(byteResult, &result)

    query, err := gojq.Parse(jq)
    if err != nil {
      fmt.Println(err)
    }
    iter := query.Run(result) // or query.RunWithContext
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

func handleRequest(req *http.Request, err error) {

}
