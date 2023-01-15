package config

import (
	"github.com/mehanizm/airtable"
	"github.com/pkg/errors"
	"os"
)

const BaseName = "Purrfect Creations (Copy)"

type Config struct {
	AirtableClient *airtable.Client
	Base           *airtable.Base
	ApiToken       string
}

func NewConfig() (*Config, error) {
	cfg := &Config{}
	apiToken := os.Getenv("AIRTABLE_API_TOKEN")
	client := airtable.NewClient(apiToken)
	cfg.AirtableClient = client
	bases, err := client.GetBases().WithOffset("").Do()
	if err != nil {
		return cfg, errors.Wrap(err, "Could not get bases")
	}

	for _, base := range bases.Bases {
		if base.Name == BaseName {
			cfg.Base = base
			break
		}
	}
	return cfg, nil
}
