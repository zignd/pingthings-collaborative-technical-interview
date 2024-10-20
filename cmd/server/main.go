package main

import (
	"github.com/rs/zerolog/log"
	"github.com/zignd/pingthings-collaborative-technical-interview/api"
	"github.com/zignd/pingthings-collaborative-technical-interview/config"
	"github.com/zignd/pingthings-collaborative-technical-interview/dependency"
)

func main() {
	cont, err := dependency.SetupContainer()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to setup DI")
	}

	var envVars *config.EnvVars
	if err := cont.Resolve(&envVars); err != nil {
		log.Fatal().Err(err).Msg("failed to resolve config.EnvVars")
	}

	app, err := api.SetupServer(cont)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to build app")
	}

	log.Info().Msgf("starting server at %s", envVars.API.Address)
	app.Listen(envVars.API.Address)
}
