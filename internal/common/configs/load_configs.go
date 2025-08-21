package configs

import (
	"github.com/go-faster/errors"
	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

const (
	EnvPath                  = ".env"
	ErrLoadCfgFile           = "failed to load dot-env file"
	ErrUnmarshalCfgsFromFile = "failed to load configurations"
)

func loadCfg(c *Config) error {
	var errLoad error
	if err := godotenv.Load(EnvPath); err != nil {
		errLoad = errors.Wrap(err, ErrLoadCfgFile) // in case with docker-compose, no need to load
	}

	if err := envconfig.Process("", c); err != nil {
		return errors.Join(errors.Wrap(err, ErrUnmarshalCfgsFromFile), errLoad)
	}

	return nil
}
