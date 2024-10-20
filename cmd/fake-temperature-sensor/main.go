package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/go-faker/faker/v4"
	"github.com/go-resty/resty/v2"
	"github.com/zignd/pingthings-collaborative-technical-interview/api"
)

const (
	APIAddress = "http://localhost:3000"
)

func main() {
	sensorName := faker.Word()

	httpClient := resty.New().
		SetBaseURL(APIAddress)

	fmt.Println("Creating sensor...")

	var sensor api.Sensor
	resp, err := httpClient.R().
		SetResult(&sensor).
		SetBody(api.Sensor{
			Name: sensorName,
			Location: api.Location{
				Longitude: faker.Longitude(),
				Latitude:  faker.Latitude(),
			},
			Tags: []string{faker.Word(), faker.Word()},
		}).
		Post("/sensors")
	if err != nil {
		panic(fmt.Errorf("failed to create sensor: %w", err))
	}
	if resp.IsError() {
		panic(fmt.Errorf("error returned by the API: %s", resp.Status()))
	}

	fmt.Printf("Sensor created: %+v\n", sensor)

	for {
		fmt.Println("Posting measurements...")

		// Generate a random temperature between 15 and 45 degrees Celsius
		randomTemperature := 15 + rand.Float64()*(45-15)

		var measurement api.Measurement
		resp, err := httpClient.R().
			SetResult(&measurement).
			SetBody(api.Measurement{
				Name:     "temperature",
				SensorID: sensor.ID,
				Value:    randomTemperature,
				Unit:     "Celsius",
			}).
			Post(fmt.Sprintf("/sensors/%s/measurements", sensor.ID))
		if err != nil {
			panic(fmt.Errorf("failed to post measurement: %w", err))
		}
		if resp.IsError() {
			panic(fmt.Errorf("error returned by the API: %s", resp.Status()))
		}

		fmt.Printf("Measurement posted: %+v\n", measurement)

		time.Sleep(100 * time.Millisecond)
	}
}
