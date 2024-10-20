package dependency

import (
	"github.com/golobby/container/v3"
	"github.com/zignd/pingthings-collaborative-technical-interview/config"
	"github.com/zignd/pingthings-collaborative-technical-interview/repository"
)

func SetupContainer() (*container.Container, error) {
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
