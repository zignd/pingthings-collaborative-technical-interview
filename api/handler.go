package api

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
	"github.com/zignd/pingthings-collaborative-technical-interview/repository"
)

func PostSensor(sensorsRepository *repository.SensorsRepository) func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		var sensor Sensor
		if err := c.BodyParser(&sensor); err != nil {
			log.Warn().Err(err).Msg("invalid request body")
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "invalid request body",
			})
		}

		ctx := c.UserContext()
		err := sensor.ValidateWithContext(ctx)
		if err != nil {
			log.Warn().Err(err).Msg("invalid sensor")
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   "invalid sensor",
				"details": err,
			})
		}

		dbSensor := &repository.Sensor{
			Name: sensor.Name,
			Location: repository.GeoJSONPoint{
				Type:        "Point",
				Coordinates: []float64{sensor.Location.Longitude, sensor.Location.Latitude},
			},
			Tags: sensor.Tags,
		}

		if err := sensorsRepository.CreateSensor(ctx, dbSensor); err != nil {
			log.Error().Err(err).Msg("failed to create sensor")
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "failed to create sensor",
			})
		}

		c.Status(fiber.StatusCreated)
		return c.JSON(mapDBSensorToAPISensor(dbSensor))
	}
}

func GetSensorByID(sensorsRepository *repository.SensorsRepository) func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		dbSensor, err := sensorsRepository.GetSensorByID(c.UserContext(), c.Params("id"))
		if err != nil {
			log.Error().Err(err).Msg("failed to get sensor")
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "failed to get sensor",
			})
		}

		return c.JSON(mapDBSensorToAPISensor(dbSensor))
	}
}

func GetSensorByName(sensorsRepository *repository.SensorsRepository) func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		dbSensor, err := sensorsRepository.GetSensorByName(c.UserContext(), c.Params("name"))
		if err != nil {
			log.Error().Err(err).Msg("failed to get sensor")
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "failed to get sensor",
			})
		}

		return c.JSON(mapDBSensorToAPISensor(dbSensor))
	}
}

func GetNearestSensor(sensorsRepository *repository.SensorsRepository) func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		latitude := c.Query("latitude")
		float64Latitude, err := strconv.ParseFloat(latitude, 64)
		if err != nil {
			log.Error().Err(err).Msg("failed to parse latitude")
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "failed to parse latitude",
			})
		}

		longitude := c.Query("longitude")
		float64Longitude, err := strconv.ParseFloat(longitude, 64)
		if err != nil {
			log.Error().Err(err).Msg("failed to parse longitude")
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "failed to parse longitude",
			})
		}

		maxDistance := c.Query("maxDistance")
		float64MaxDistance, err := strconv.ParseFloat(maxDistance, 64)
		if err != nil {
			log.Error().Err(err).Msg("failed to parse maxDistance")
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "failed to parse maxDistance",
			})
		}

		dbSensor, err := sensorsRepository.GetNearestSensor(c.UserContext(), float64Latitude, float64Longitude, float64MaxDistance)
		if err != nil {
			log.Error().Err(err).Msg("failed to get nearest sensor")
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "failed to get nearest sensor",
			})
		}

		if dbSensor == nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "no sensor found within the specified distance",
			})
		}

		return c.JSON(mapDBSensorToAPISensor(dbSensor))
	}
}

func PutSensor(sensorsRepository *repository.SensorsRepository) func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		var sensor Sensor
		if err := c.BodyParser(&sensor); err != nil {
			log.Warn().Err(err).Msg("invalid request body")
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "invalid request body",
			})
		}

		ctx := c.UserContext()
		err := sensor.ValidateWithContext(ctx)
		if err != nil {
			log.Warn().Err(err).Msg("invalid sensor")
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   "invalid sensor",
				"details": err,
			})
		}

		dbSensor := &repository.Sensor{
			Name: sensor.Name,
			Location: repository.GeoJSONPoint{
				Type:        "Point",
				Coordinates: []float64{sensor.Location.Longitude, sensor.Location.Latitude},
			},
			Tags: sensor.Tags,
		}

		if err := sensorsRepository.UpdateSensor(ctx, c.Params("id"), dbSensor); err != nil {
			log.Error().Err(err).Msg("failed to update sensor")
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "failed to update sensor",
			})
		}

		return c.JSON(mapDBSensorToAPISensor(dbSensor))
	}
}

func PostMeasurement(sensorsRepository *repository.SensorsRepository, measurementRepository *repository.MeasurementRepository) func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		var measurement Measurement
		if err := c.BodyParser(&measurement); err != nil {
			log.Warn().Err(err).Msg("invalid request body")
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "invalid request body",
			})
		}

		ctx := c.UserContext()
		err := measurement.ValidateWithContext(ctx)
		if err != nil {
			log.Warn().Err(err).Msg("invalid measurement")
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   "invalid measurement",
				"details": err,
			})
		}

		sensor, err := sensorsRepository.GetSensorByID(ctx, c.Params("id"))
		if err != nil {
			log.Error().Err(err).Msg("failed to get sensor")
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "failed to get sensor",
			})
		}
		if sensor == nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "sensor not found",
			})
		}

		dbMeasurement := mapAPIMeasurementToDBMeasurement(&measurement)
		dbMeasurement.SensorID = sensor.ID.Hex()

		if err := measurementRepository.CreateMeasurement(ctx, dbMeasurement); err != nil {
			log.Error().Err(err).Msg("failed to create measurement")
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "failed to create measurement",
			})
		}

		measurement.SensorID = dbMeasurement.SensorID
		measurement.Timestamp = dbMeasurement.Timestamp

		c.Status(fiber.StatusCreated)
		return c.JSON(measurement)
	}
}

func GetMeasurementSummary(sensorsRepository *repository.SensorsRepository, measurementRepository *repository.MeasurementRepository) func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		sensor, err := sensorsRepository.GetSensorByID(c.UserContext(), c.Params("id"))
		if err != nil {
			log.Error().Err(err).Msg("failed to get sensor")
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "failed to get sensor",
			})
		}
		if sensor == nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "sensor not found",
			})
		}

		measurement := c.Query("measurement")
		if measurement == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "measurement query parameter is required",
			})
		}

		unit := c.Query("unit")
		if unit == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "unit query parameter is required",
			})
		}

		start := c.Query("start")
		end := c.Query("end")
		if start == "" || end == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "start and end query parameters are required",
			})
		}

		startTime, err := time.Parse(time.RFC3339, start)
		if err != nil {
			log.Error().Err(err).Msg("failed to parse start query parameter")
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "failed to parse start query parameter",
			})
		}

		endTime, err := time.Parse(time.RFC3339, end)
		if err != nil {
			log.Error().Err(err).Msg("failed to parse end query parameter")
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "failed to parse end query parameter",
			})
		}

		summary, err := measurementRepository.GetMeasurementSummary(c.UserContext(), sensor.ID.Hex(), measurement, unit, startTime, endTime)
		if err != nil {
			log.Error().Err(err).Msg("failed to get measurement summary")
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "failed to get measurement summary",
			})
		}
		if summary.Count == 0 {
			return c.JSON(fiber.Map{
				"message": "no measurements found for the specified time range",
			})
		}

		return c.JSON(summary)
	}
}
