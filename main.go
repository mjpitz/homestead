package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/mjpitz/homestead/apis/geocoding"
	"github.com/mjpitz/homestead/apis/weather"
)

type Config struct {
	Address *geocoding.Address `json:"address,omitempty"`
}

func main() {
	body, err := ioutil.ReadFile("config.json")
	if err != nil {
		panic(err)
	}

	config := &Config{}
	err = json.NewDecoder(bytes.NewReader(body)).Decode(config)
	if err != nil {
		panic(err)
	}

	ctx := context.Background()
	geocodingAPI := geocoding.NewClient()
	weatherAPI := weather.NewClient()

	geocodeResp, err := geocodingAPI.SearchByAddress(ctx, config.Address)
	if err != nil {
		panic(err)
	}

	coordinates := geocodeResp.Result.AddressMatches[0].Coordinates

	point, err := weatherAPI.GetPoint(ctx, coordinates.Y, coordinates.X)
	if err != nil {
		panic(err)
	}

	gridpoint, err := weatherAPI.GetGridpoint(ctx, point.GridID, point.GridX, point.GridY)
	if err != nil {
		panic(err)
	}

	fmt.Println(gridpoint)
}
