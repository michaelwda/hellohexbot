package main

import (
	"encoding/json"
	"fmt"
	"image"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
	"github.com/lucasb-eyer/go-colorful"
	"golang.org/x/image/colornames"
)

const (
	url               = "https://api.noopschallenge.com/hexbot"
	height            = 400
	width             = 400
	minPixelWidth     = 5
	minPixelHeight    = 5
	maxPixelWidth     = 5
	maxPixelHeight    = 5
	apiColorCount     = 200
	snapPixelToGrid   = true
	gridInterval      = 5
	requestsPerSecond = 20
)

type colorUpdate struct {
	hex []string
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

	//on-tick- add a value to the channel - only X requests a second
	go func() {
		milliseconds := 1000 / requestsPerSecond
		for t := range time.Tick(time.Duration(milliseconds) * time.Millisecond) {
			limiter <- t
		}
	}()

	//make a buffered channel of 100 updates
	updates := make(chan colorUpdate, 100)

	//spawn 50 goroutines, passing in my rate limiter
	for i := 0; i < 50; i++ {
		go callAPI(updates, limiter)
	}

	for !win.Closed() {
		win.Update()

		//read an update
		select {
		case update := <-updates:
			{
				for i := 0; i < len(update.hex); i++ {
					x := rand.Intn(width + gridInterval)
					y := rand.Intn(height + gridInterval)

					if snapPixelToGrid {
						//use int division to drop remainders
						x = (x / gridInterval) * gridInterval
						y = (y / gridInterval) * gridInterval
					}

					w := minPixelWidth
					h := minPixelHeight
					if maxPixelWidth > minPixelWidth {
						w = rand.Intn(maxPixelWidth-minPixelWidth) + minPixelWidth
					}
					if maxPixelHeight > minPixelHeight {
						h = rand.Intn(maxPixelHeight-minPixelHeight) + minPixelHeight
					}

					sprite := createSpriteFromHex(update.hex[i], w, h)
					sprite.Draw(win, pixel.IM.Moved(pixel.V(float64(x), float64(y))))
				}
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

func callAPI(updates chan colorUpdate, limiter chan time.Time) {
	//repeatedly call the api and get colors, with respect to the limiter
	for true {
		<-limiter
		response, err := http.Get(url + "?count=" + strconv.Itoa(apiColorCount))
		if err == nil {
			body, _ := ioutil.ReadAll(response.Body)
			hex := getHex(body)
			update := colorUpdate{
				hex: hex,
			}
			updates <- update
		}
	}
}

func getHex(body []byte) []string {

	var hex []string

	var data map[string]interface{}
	//var assignment not allowed inside an if. must use := shorthand
	//inlining the nil check using chaining
	if err := json.Unmarshal(body, &data); err != nil {
		hex = append(hex, "#000000")
	}
	colors := data["colors"].([]interface{})
	for i := 0; i < len(colors); i++ {
		color := colors[i].(map[string]interface{})
		hex = append(hex, color["value"].(string))
	}
	return hex
}
