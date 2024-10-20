package api

import (
	"github.com/gofiber/contrib/fiberzerolog"
	"github.com/gofiber/fiber/v2"
	"github.com/golobby/container/v3"
	"github.com/rs/zerolog"
	"github.com/zignd/pingthings-collaborative-technical-interview/repository"
)

func SetupServer(cont *container.Container) (*fiber.App, error) {
	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
	})

	err := cont.Call(func(
		logger zerolog.Logger,
		sensorsRepository *repository.SensorsRepository,
		measurementRepository *repository.MeasurementRepository,
	) {
		app.Use(fiberzerolog.New(fiberzerolog.Config{
			Logger:   &logger,
			Messages: []string{"server side error", "client side error", "success"},
		}))

		app.Post("/sensors", PostSensor(sensorsRepository))
		app.Get("/sensors/nearest", GetNearestSensor(sensorsRepository))
		app.Get("/sensors/name/:name", GetSensorByName(sensorsRepository))
		app.Get("/sensors/:id", GetSensorByID(sensorsRepository))
		app.Put("/sensors/:id", PutSensor(sensorsRepository))
		app.Post("/sensors/:id/measurements", PostMeasurement(sensorsRepository, measurementRepository))
		app.Get("/sensors/:id/measurements/summary", GetMeasurementSummary(sensorsRepository, measurementRepository))
	})
	if err != nil {
		return nil, err
	}

	return app, nil
}
