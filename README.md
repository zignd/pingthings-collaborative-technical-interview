# pingthings-collaborative-technical-interview

This is a collaborative technical interview project for PingThings. The project is a RESTful API that allows the management of sensors and their measurements. The API is built using Go and the Fiber framework. The data is stored in MongoDB and InfluxDB. MongoDB is used to store the sensor data and InfluxDB is used to store the measurements data.

## Setup Environment

### Prerequisites
Ensure you have the following installed:
- `go 1.23+`
- `make`
- `docker`
- `docker-compose`

### Steps to Setup the Environment

The project depends on a MongoDB and an InfluxDB which are both provided as docker containers. To start the containers, run the following command:
```bash
make deps-up
```

To stop the containers, run the following command:
```bash
make deps-down
```

### Running the tests

Requirements:
* Ensure the dependencies are running following the instructions in the section "Steps to Setup the Environment".

To run the tests, run the following command:
```bash
make tests
```

### Running the application

Requirements:
* Ensure the dependencies are running following the instructions in the section "Steps to Setup the Environment".

To run the application, run the following command:
```bash
make run-api
```

The application will be running on `http://localhost:3000` by default based on the configuration in the `.env` file.

There's also a fake sensor that sends data to the application. To start the fake sensor, run the following command:
```bash
make run-fake-temperature-sensor
```

### API Documentation

#### POST /sensors

Example:
```
curl --location 'http://localhost:3000/sensors' \
--header 'Content-Type: application/json' \
--data '{
    "name": "farm-1",
    "location": {
        "longitude": 100.425423575758686,
        "latitude": -50.26740787628066
    },
    "tags": [
        "tag1",
        "tag2"
    ]
}'
```

#### POST /sensors/:id/measurements

Example:
```
curl --location 'http://localhost:3000/sensors/6717bedc52536d1a81f9fca7/measurements' \
--header 'Content-Type: application/json' \
--data '{
  "name": "temperature",
	"unit": "celsius",
	"value": 16.4
}'
```

#### GET /sensors/:id/measurements/summary?start=:start&end=:end&measurement=:measurement&unit=:unit

Example:
```
curl --location 'http://localhost:3000/sensors/6717bedc52536d1a81f9fca7/measurements/summary?start=2021-05-03T00%3A00%3A00Z&end=2024-10-30T15%3A00%3A00Z&measurement=temperature&unit=celsius'
```

#### PUT /sensors/:id

Example:
```
curl --location --request PUT 'http://localhost:3000/sensors/6717bedc52536d1a81f9fca7' \
--header 'Content-Type: application/json' \
--data '{
    "name": "farm-old",
    "location": {
        "longitude": -25.425423575758686,
        "latitude": -49.26740787628066
    },
    "tags": [
        "tag1",
        "tag2",
        "tag3"
    ]
}'
```

#### PUT /sensors/:id

Example:
```
curl --location --request PUT 'http://localhost:3000/sensors/6717bedc52536d1a81f9fca7' \
--header 'Content-Type: application/json' \
--data '{
    "name": "farm-old",
    "location": {
        "longitude": -25.425423575758686,
        "latitude": -49.26740787628066
    },
    "tags": [
        "tag1",
        "tag2",
        "tag3"
    ]
}'
```

#### GET /sensor/:name

Example:
```
curl --location 'http://localhost:3000/sensors/name/farm-old'
```

#### GET /sensor/nearest?longitude=:longitude&latitude=:latitude&maxDistance=:maxDistance

Example:
```
curl --location 'http://localhost:3000/sensors/nearest?longitude=-25&latitude=-50&maxDistance=100000'
```