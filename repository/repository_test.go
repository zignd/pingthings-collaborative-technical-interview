package repository

import (
	"context"
	"fmt"
	"math/rand/v2"
	"testing"
	"time"

	"github.com/go-faker/faker/v4"
	"github.com/golobby/container/v3"
	"github.com/stretchr/testify/require"
	"github.com/zignd/pingthings-collaborative-technical-interview/config"
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

func setupContainer() (*container.Container, error) {
	cont := container.New()

	if err := cont.Singleton(config.Get); err != nil {
		return nil, err
	}
	if err := cont.Singleton(buildMongoClient); err != nil {
		return nil, err
	}
	if err := cont.Singleton(NewSensorsRepository); err != nil {
		return nil, err
	}
	if err := cont.Singleton(NewMeasurementRepository); err != nil {
		return nil, err
	}

	return &cont, nil
}

func TestSensorsRepository(t *testing.T) {
	t.Parallel()
	is := require.New(t)

	cont, err := setupContainer()
	is.Nil(err)

	var envVars *config.EnvVars
	is.Nil(cont.Resolve(&envVars))

	var mongoClient *mongo.Client
	is.Nil(cont.Resolve(&mongoClient))

	sensorsRepository, err := NewSensorsRepository(envVars, mongoClient)
	is.Nil(err)

	ctx := context.Background()

	t.Run("when CreateSensor is invoked with a valid sensor, it should create that sensor in the database", func(t *testing.T) {
		t.Parallel()
		is := require.New(t)

		newSensor := &Sensor{
			Name: "Sensor 1",
			Location: GeoJSONPoint{
				Type:        "Point",
				Coordinates: []float64{1.0, 1.0},
			},
			Tags: []string{"tag1", "tag2"},
		}
		err = sensorsRepository.CreateSensor(ctx, newSensor)
		is.Nil(err)
		is.False(newSensor.ID.IsZero())
	})

	t.Run("when GetSensorByID is invoke with a valid sensor ID, it should return that sensor", func(t *testing.T) {
		t.Parallel()
		is := require.New(t)

		newSensor := &Sensor{
			Name: "Sensor 2",
			Location: GeoJSONPoint{
				Type:        "Point",
				Coordinates: []float64{2.0, 2.0},
			},
			Tags: []string{"tag3", "tag4"},
		}
		err = sensorsRepository.CreateSensor(ctx, newSensor)
		is.Nil(err)

		foundSensor, err := sensorsRepository.GetSensorByID(ctx, newSensor.ID.Hex())
		is.Nil(err)

		is.Equal(newSensor.ID, foundSensor.ID)
	})
}

func TestMeasurementRepository(t *testing.T) {
	t.Parallel()
	is := require.New(t)

	cont, err := setupContainer()
	is.Nil(err)

	var envVars *config.EnvVars
	is.Nil(cont.Resolve(&envVars))

	var measurementRepository *MeasurementRepository
	is.Nil(cont.Resolve(&measurementRepository))

	ctx := context.Background()

	t.Run("when CreateMeasurement is invoked with a valid measurement, it should create that measurement in the database", func(t *testing.T) {
		t.Parallel()
		is := require.New(t)

		newMeasurement := &Measurement{
			Name:     "temperature",
			SensorID: faker.UUIDHyphenated(),
			Unit:     "celsius",
			Value:    15.3,
		}
		err = measurementRepository.CreateMeasurement(ctx, newMeasurement)
		is.Nil(err)
		is.NotZero(newMeasurement.Timestamp)
	})

	t.Run("when GetMeasurementSummary is invoked with a valid sensor ID, start, and end time, it should return the summary of the measurements", func(t *testing.T) {
		t.Parallel()
		is := require.New(t)

		sensorID := faker.UUIDHyphenated()
		measurementName := "temperature"
		measurementUnit := "celsius"
		measurementsCount := 10
		for i := 0; i < measurementsCount; i++ {
			randomTemperature := 15 + rand.Float64()*(45-15)
			newMeasurement := &Measurement{
				Name:     "temperature",
				SensorID: sensorID,
				Unit:     measurementUnit,
				Value:    randomTemperature,
			}
			err = measurementRepository.CreateMeasurement(ctx, newMeasurement)
			is.Nil(err)
		}

		start := time.Now().Add(-1 * time.Hour)
		end := time.Now().Add(1 * time.Hour)

		summary, err := measurementRepository.GetMeasurementSummary(ctx, sensorID, measurementName, measurementUnit, start, end)
		is.Nil(err)

		is.GreaterOrEqual(summary.MinValue, 15.0)
		is.LessOrEqual(summary.MaxValue, 45.0)
		is.True(summary.MedianValue >= 15 && summary.MedianValue <= 45)
		is.True(summary.MeanValue >= 15 && summary.MeanValue <= 45)
		is.Equal(measurementUnit, summary.Unit)
		is.Equal(measurementsCount, summary.Count)
	})

	t.Run("when GetMeasurementSummary is invoked with an invalid sensor ID, it should return an error", func(t *testing.T) {
		t.Parallel()
		is := require.New(t)

		sensorID := faker.UUIDHyphenated()

		start := time.Now().Add(-1 * time.Hour)
		end := time.Now().Add(1 * time.Hour)

		_, err := measurementRepository.GetMeasurementSummary(ctx, sensorID, "temperature", "celsius", end, start)
		is.ErrorContains(err, "failed to query measurement summary: invalid: error in building plan while starting program: cannot query an empty range")
	})
}
