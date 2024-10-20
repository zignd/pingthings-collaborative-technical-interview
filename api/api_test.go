package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/golobby/container/v3"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/require"
	"github.com/zignd/pingthings-collaborative-technical-interview/config"
	"github.com/zignd/pingthings-collaborative-technical-interview/repository"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func buildMongoClient(envVars *config.EnvVars) (*mongo.Client, error) {
	ctx := context.Background()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(envVars.MongoDB.URI))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}
	return client, nil
}

func configureLogger(envVars *config.EnvVars) zerolog.Logger {
	return log.Logger
}

func setupContainer() (*container.Container, error) {
	cont := container.New()

	if err := cont.Singleton(config.Get); err != nil {
		return nil, err
	}
	if err := cont.Singleton(configureLogger); err != nil {
		return nil, err
	}
	if err := cont.Singleton(buildMongoClient); err != nil {
		return nil, err
	}
	if err := cont.Singleton(repository.NewSensorsRepository); err != nil {
		return nil, err
	}
	if err := cont.Singleton(repository.NewMeasurementRepository); err != nil {
		return nil, err
	}

	return &cont, nil
}

func TestAPI(t *testing.T) {
	t.Parallel()
	is := require.New(t)

	cont, err := setupContainer()
	is.Nil(err)

	app, err := SetupServer(cont)
	is.Nil(err)

	t.Run("when a new sensor is created, it should have a unique ID", func(t *testing.T) {
		t.Parallel()
		is := require.New(t)

		ctx := context.Background()

		body := Sensor{
			Name: faker.Word(),
			Location: Location{
				Longitude: faker.Longitude(),
				Latitude:  faker.Latitude(),
			},
			Tags: []string{faker.Word(), faker.Word()},
		}
		bodyBytes, err := json.Marshal(body)
		is.Nil(err)

		req := httptest.NewRequestWithContext(ctx, "POST", "/sensors", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")

		res, err := app.Test(req)
		is.Nil(err)
		is.Equal(http.StatusCreated, res.StatusCode)
	})

	t.Run("when a sensor is created and then retrieved by ID, it should be the same sensor", func(t *testing.T) {
		t.Parallel()
		is := require.New(t)

		ctx := context.Background()

		body := Sensor{
			Name: faker.Word(),
			Location: Location{
				Longitude: faker.Longitude(),
				Latitude:  faker.Latitude(),
			},
			Tags: []string{faker.Word(), faker.Word()},
		}
		bodyBytes, err := json.Marshal(body)
		is.Nil(err)

		req := httptest.NewRequestWithContext(ctx, "POST", "/sensors", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")

		res, err := app.Test(req)
		is.Nil(err)
		is.Equal(http.StatusCreated, res.StatusCode)

		var createdSensor Sensor
		err = json.NewDecoder(res.Body).Decode(&createdSensor)
		is.Nil(err)

		req = httptest.NewRequestWithContext(ctx, "GET", fmt.Sprintf("/sensors/%s", createdSensor.ID), nil)
		res, err = app.Test(req)
		is.Nil(err)
		is.Equal(http.StatusOK, res.StatusCode)

		var retrievedSensor Sensor
		err = json.NewDecoder(res.Body).Decode(&retrievedSensor)
		is.Nil(err)

		is.Equal(createdSensor, retrievedSensor)
	})

	t.Run("when the sensor contract is violated, it should return a bad request", func(t *testing.T) {
		t.Parallel()
		is := require.New(t)

		ctx := context.Background()

		body := Sensor{
			// Empty on purpose
		}
		bodyBytes, err := json.Marshal(body)
		is.Nil(err)

		req := httptest.NewRequestWithContext(ctx, "POST", "/sensors", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")

		res, err := app.Test(req)
		is.Nil(err)
		is.Equal(http.StatusBadRequest, res.StatusCode)

		resBody := map[string]interface{}{}
		json.NewDecoder(res.Body).Decode(&resBody)
		expectedResBody := map[string]interface{}{
			"details": map[string]interface{}{
				"location": map[string]interface{}{
					"latitude":  "cannot be blank",
					"longitude": "cannot be blank",
				},
				"name": "cannot be blank",
				"tags": "cannot be blank",
			},
			"error": "invalid sensor",
		}
		is.Equal(expectedResBody, resBody)
	})
}
