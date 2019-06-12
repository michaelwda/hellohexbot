package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
	"time"
)

const url = "https://api.noopschallenge.com/hexbot"

func main() {

	fmt.Printf("Hello HexBot!\n")

	//create a waitgroup to sync my network goroutines
	var wg sync.WaitGroup

	//make a buffered channel of 100 strings
	messages := make(chan string, 100)

	//make a rate-limiter
	//1 request per 50 milliseconds, 20 requests a second, our 100 requests should take about 5 seconds to complete
	limiter := make(chan time.Time, 1)

	//on-tick- add a value to the channel
	go func() {
		for t := range time.Tick(50 * time.Millisecond) {
			limiter <- t
		}
	}()

	//spawn 100 goroutines, passing in my rate limiter and the syncgroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go CallApi(messages, limiter, &wg)
	}

	//wait for all routines to return
	wg.Wait()

	//receive all from channel
	for i := 0; i < 100; i++ {
		msg := <-messages
		fmt.Println(msg)
	}

}

func CallApi(messages chan string, limiter chan time.Time, wg *sync.WaitGroup) {
	<-limiter
	fmt.Println("request", time.Now())
	response, err := http.Get(url)
	if err != nil {
		messages <- err.Error()
	} else {
		body, _ := ioutil.ReadAll(response.Body)
		messages <- GetHex(body)
	}
	wg.Done()
}

func GetHex(body []byte) string {
	var data map[string]interface{}
	//var assignment not allowed inside an if. must use := shorthand
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
