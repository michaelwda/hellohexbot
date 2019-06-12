package main

import (
	"encoding/json"
	"fmt"
	"image"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
	"github.com/lucasb-eyer/go-colorful"
	"golang.org/x/image/colornames"
)

const (
	url         = "https://api.noopschallenge.com/hexbot"
	height      = 400
	width       = 400
	pixelWidth  = 50
	pixelHeight = 50
)

type colorUpdate struct {
	x   int
	y   int
	w   int
	h   int
	hex string
}

func main() {
	fmt.Printf("Hello HexBot!\n")
	pixelgl.Run(run)
}

func run() {

	//set up our window
	cfg := pixelgl.WindowConfig{
		Title:  "HexBot!",
		Bounds: pixel.R(0, 0, width, height),
		VSync:  true,
	}
	win, err := pixelgl.NewWindow(cfg)
	if err != nil {
		panic(err)
	}

	//initially make it blue
	win.Clear(colornames.Skyblue)

	//make a rate-limiter
	limiter := make(chan time.Time, 1)

	//on-tick- add a value to the channel - only 10 requests a second
	go func() {
		for t := range time.Tick(100 * time.Millisecond) {
			limiter <- t
		}
	}()

	//make a buffered channel of 100 updates
	updates := make(chan colorUpdate, 100)

	//spawn 50 goroutines, passing in my rate limiter
	for i := 0; i < 50; i++ {
		go callApi(updates, limiter)
	}

	for !win.Closed() {
		win.Update()

		//read an update
		select {
		case update := <-updates:
			{
				fmt.Println("received update")
				sprite := createSpriteFromHex(update.hex, update.w, update.h)
				sprite.Draw(win, pixel.IM.Moved(pixel.V(float64(update.x), float64(update.y))))
			}
		default:
			{
				//do nothing, continue to render loop
			}
		}
	}
}

func createSpriteFromHex(hex string, w int, h int) *pixel.Sprite {
	c, err := colorful.Hex(hex)
	if err != nil {
		log.Fatal(err)
	}
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for i := 0; i < w; i++ {
		for j := 0; j < h; j++ {
			img.Set(i, j, c)
		}
	}
	pic := pixel.PictureDataFromImage(img)
	sprite := pixel.NewSprite(pic, pic.Bounds())

	return sprite
}

func callApi(updates chan colorUpdate, limiter chan time.Time) {
	//repeatedly call the api and get colors, with respect to the limiter
	for true {
		<-limiter
		response, err := http.Get(url)
		if err == nil {
			body, _ := ioutil.ReadAll(response.Body)
			update := colorUpdate{
				x:   rand.Intn(width),
				y:   rand.Intn(height),
				w:   rand.Intn(pixelWidth-1) + 1,
				h:   rand.Intn(pixelHeight-1) + 1,
				hex: getHex(body),
			}
			updates <- update
		}
	}
}

func getHex(body []byte) string {
	var data map[string]interface{}
	//var assignment not allowed inside an if. must use := shorthand
	//inlining the nil check using chaining
	if err := json.Unmarshal(body, &data); err != nil {
		return "#000000"
	}
	colors := data["colors"].([]interface{})
	firstColor := colors[0].(map[string]interface{})
	return firstColor["value"].(string)
}
