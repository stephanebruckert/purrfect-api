package config

import (
	"github.com/caarlos0/env"
	"github.com/pkg/errors"
)

type Config struct {
	ApiToken  string `env:"AIRTABLE_API_TOKEN" envDefault:""`
	SmeeURL   string `env:"SMEE_URL" envDefault:"https://smee.io/2mxhU4Pb2YrNvF8E"`
	BaseName  string `env:"BASE_NAME" envDefault:"Purrfect Creations"`
	TableName string `env:"TABLE_NAME" envDefault:"Orders"`
}

func NewConfig() (*Config, error) {
	var config Config

	if err := env.Parse(&config); err != nil {
		return nil, errors.Wrap(err, "error initializing config")
	}

	return &config, nil
}
