package config

import (
	"github.com/Netflix/go-env"
	"github.com/joho/godotenv"
)

type EnvVars struct {
	API struct {
		Address string `env:"API__ADDRESS,required=true"`
	}
	MongoDB struct {
		URI      string `env:"MONGODB__URI,required=true"`
		Database string `env:"MONGODB__DATABASE,required=true"`
	}
	InfluxDB struct {
		ServerURL string `env:"INFLUXDB__SERVER_URL,required=true"`
		Org       string `env:"INFLUXDB__ORG,required=true"`
		Bucket    string `env:"INFLUXDB__BUCKET,required=true"`
		Token     string `env:"INFLUXDB__TOKEN,required=true"`
	}
	DevMode bool `env:"DEV_MODE"`
}

func Get() (*EnvVars, error) {
	_ = godotenv.Load()
	var envVars EnvVars
	if _, err := env.UnmarshalFromEnviron(&envVars); err != nil {
		return nil, err
	}
	return &envVars, nil
}
