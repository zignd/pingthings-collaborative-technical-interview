.PHONY: deps-up deps-down run-api run-fake-temperature-sensor tests

# Ensuring the .env file exists
ifeq (,$(wildcard .env))
$(error .env file not found)
endif

# Loading the environment variables
include .env
export $(shell sed 's/=.*//' .env)

deps-up:
	docker compose up

deps-down:
	docker compose down

run-api:
	go run cmd/server/main.go

run-fake-temperature-sensor:
	go run cmd/fake-temperature-sensor/main.go

tests:
	go test -v -timeout 10s ./...