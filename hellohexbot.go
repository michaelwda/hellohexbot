package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

const url = "https://api.noopschallenge.com/hexbot"

func main() {
	fmt.Printf("Hello HexBot!\n")
	response, err := http.Get(url)
	if err != nil {
		fmt.Printf("The HTTP request failed with error %s\n", err)
	} else {
		body, _ := ioutil.ReadAll(response.Body)
		hex := GetHex(body)
		fmt.Printf("My hex value is: %s", hex)
	}

}

func GetHex(body []byte) string {

	var data map[string]interface{}
	//var assignmnt not allowed inside an if. must use := shorthand
	//inlining the nil check using chaining
	//interestingly, if you don't use K&R bracket style here it gives a syntax error?
	if err := json.Unmarshal(body, &data); err != nil {
		fmt.Printf("The HTTP request failed with error %s\n", err)
		panic(err)
	}
	colors := data["colors"].([]interface{})
	firstColor := colors[0].(map[string]interface{})
	return firstColor["value"].(string)
}
