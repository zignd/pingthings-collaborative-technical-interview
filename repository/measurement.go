package repository

import (
	"context"
	"fmt"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	influxdb2api "github.com/influxdata/influxdb-client-go/v2/api"
	"github.com/rs/zerolog/log"
	"github.com/zignd/pingthings-collaborative-technical-interview/config"
)

type Measurement struct {
	Name      string
	SensorID  string
	Unit      string
	Value     float64
	Timestamp time.Time
}

type MeasurementSummary struct {
	MinValue    float64 `json:"min_value"`
	MaxValue    float64 `json:"max_value"`
	MedianValue float64 `json:"average_value"`
	MeanValue   float64 `json:"mean_value"`
	Unit        string  `json:"unit"`
	Count       int     `json:"count"`
}

type MeasurementRepository struct {
	client   influxdb2.Client
	bucket   string
	writeAPI influxdb2api.WriteAPIBlocking
	queryAPI influxdb2api.QueryAPI
}

func NewMeasurementRepository(envVars *config.EnvVars) *MeasurementRepository {
	m := &MeasurementRepository{
		client: influxdb2.NewClient(envVars.InfluxDB.ServerURL, envVars.InfluxDB.Token),
		bucket: envVars.InfluxDB.Bucket,
	}

	m.writeAPI = m.client.WriteAPIBlocking(envVars.InfluxDB.Org, envVars.InfluxDB.Bucket)
	m.queryAPI = m.client.QueryAPI(envVars.InfluxDB.Org)

	return m
}

func (m *MeasurementRepository) Close() error {
	m.client.Close()
	return nil
}

func (m *MeasurementRepository) CreateMeasurement(ctx context.Context, measurement *Measurement) error {
	timestamp := time.Now()

	p := influxdb2.NewPointWithMeasurement(measurement.Name).
		AddTag("unit", measurement.Unit).
		AddTag("sensor_id", measurement.SensorID).
		AddField("value", measurement.Value).
		SetTime(timestamp)
	if err := m.writeAPI.WritePoint(ctx, p); err != nil {
		return fmt.Errorf("failed to write the measurement point: %w", err)
	}

	measurement.Timestamp = timestamp

	return nil
}

func (m *MeasurementRepository) GetMeasurementSummary(ctx context.Context, sensorID, measurement, unit string, start, end time.Time) (*MeasurementSummary, error) {
	query := fmt.Sprintf(
		`result = from(bucket: "%s")
			|> range(start: %s, stop: %s)
			|> filter(fn: (r) => r["_measurement"] == "%s")
			|> filter(fn: (r) => r["sensor_id"] == "%s")
			|> filter(fn: (r) => r["unit"] == "%s")
			
		result
			|> aggregateWindow(every: 1d, fn: mean, createEmpty: false)
			|> yield(name: "mean")

		result
			|> aggregateWindow(every: 1d, fn: median, createEmpty: false)
			|> yield(name: "median")

		result
			|> count()
			|> yield(name: "count")

		result
			|> min()
			|> yield(name: "min")

		result
			|> min()
			|> yield(name: "max")`,
		m.bucket, start.Format(time.RFC3339), end.Format(time.RFC3339), measurement, sensorID, unit)

	log.Info().Str("query", query).Msg("executing query")

	result, err := m.queryAPI.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query measurement summary: %w", err)
	}

	const (
		meanResultName   = "mean"
		medianResultName = "median"
		countResultName  = "count"
		minResultName    = "min"
		maxResultName    = "max"
	)

	measurementSummary := MeasurementSummary{}

	for result.Next() {
		if result.Err() != nil {
			return nil, fmt.Errorf("query error: %w", result.Err())
		}

		if !result.TableChanged() {
			continue
		}

		if measurementSummary.Unit == "" {
			unit, ok := result.Record().ValueByKey("unit").(string)
			if !ok {
				return nil, fmt.Errorf("unexpected type for unit value: %T", result.Record().Value())
			}
			measurementSummary.Unit = unit
		}

		resultName := result.Record().Result()
		switch resultName {
		case meanResultName:
			meanValue, ok := result.Record().Value().(float64)
			if !ok {
				return nil, fmt.Errorf("unexpected type for mean value: %T", result.Record().Value())
			}
			measurementSummary.MeanValue = meanValue
		case medianResultName:
			medianValue, ok := result.Record().Value().(float64)
			if !ok {
				return nil, fmt.Errorf("unexpected type for median value: %T", result.Record().Value())
			}
			measurementSummary.MedianValue = medianValue
		case countResultName:
			count, ok := result.Record().Value().(int64)
			if !ok {
				return nil, fmt.Errorf("unexpected type for count value: %T", result.Record().Value())
			}
			measurementSummary.Count = int(count)
		case minResultName:
			minValue, ok := result.Record().Value().(float64)
			if !ok {
				return nil, fmt.Errorf("unexpected type for min value: %T", result.Record().Value())
			}
			measurementSummary.MinValue = minValue
		case maxResultName:
			maxValue, ok := result.Record().Value().(float64)
			if !ok {
				return nil, fmt.Errorf("unexpected type for max value: %T", result.Record().Value())
			}
			measurementSummary.MaxValue = maxValue
		default:
			return nil, fmt.Errorf("unexpected result name: %s", resultName)
		}
	}

	return &measurementSummary, nil
}
