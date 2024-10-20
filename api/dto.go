package api

import (
	"context"
	"time"

	validator "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/zignd/pingthings-collaborative-technical-interview/repository"
)

type Location struct {
	Longitude float64 `json:"longitude"`
	Latitude  float64 `json:"latitude"`
}

func (l Location) ValidateWithContext(ctx context.Context) error {
	fieldRules := []*validator.FieldRules{
		validator.Field(&l.Longitude, validator.Required, validator.Min(-180.0), validator.Max(180.0)),
		validator.Field(&l.Latitude, validator.Required, validator.Min(-90.0), validator.Max(90.0)),
	}

	return validator.ValidateStructWithContext(ctx, &l, fieldRules...)
}

type Sensor struct {
	ID       string   `json:"id,omitempty"`
	Name     string   `json:"name"`
	Location Location `json:"location"`
	Tags     []string `json:"tags"`
}

func (s Sensor) ValidateWithContext(ctx context.Context) error {
	fieldRules := []*validator.FieldRules{
		validator.Field(&s.Name, validator.Required),
		validator.Field(&s.Location, validator.Required),
		validator.Field(&s.Tags, validator.Required, validator.Length(1, 0)),
	}

	return validator.ValidateStructWithContext(ctx, &s, fieldRules...)
}

func mapDBSensorToAPISensor(dbSensor *repository.Sensor) *Sensor {
	return &Sensor{
		ID:   dbSensor.ID.Hex(),
		Name: dbSensor.Name,
		Location: Location{
			Longitude: dbSensor.Location.Coordinates[0],
			Latitude:  dbSensor.Location.Coordinates[1],
		},
		Tags: dbSensor.Tags,
	}
}

type Measurement struct {
	Name      string    `json:"name"`
	SensorID  string    `json:"sensor_id"`
	Unit      string    `json:"unit"`
	Value     float64   `json:"value"`
	Timestamp time.Time `json:"timestamp"`
}

func (m Measurement) ValidateWithContext(ctx context.Context) error {
	fieldRules := []*validator.FieldRules{
		validator.Field(&m.Name, validator.Required),
		validator.Field(&m.Unit, validator.Required),
		validator.Field(&m.Value, validator.Required),
	}

	return validator.ValidateStructWithContext(ctx, &m, fieldRules...)
}

func mapAPIMeasurementToDBMeasurement(apiMeasurement *Measurement) *repository.Measurement {
	return &repository.Measurement{
		Name:      apiMeasurement.Name,
		SensorID:  apiMeasurement.SensorID,
		Unit:      apiMeasurement.Unit,
		Value:     apiMeasurement.Value,
		Timestamp: apiMeasurement.Timestamp,
	}
}
