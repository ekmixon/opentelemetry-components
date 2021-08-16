package mysqlreceiver

import (
	"errors"

	"go.opentelemetry.io/collector/receiver/scraperhelper"
)

type Config struct {
	scraperhelper.ScraperControllerSettings `mapstructure:",squash"`
	Username                                string `mapstructure:"username"`
	Password                                string `mapstructure:"password"`
	Database                                string `mapstructure:"database"`
	Endpoint                                string `mapstructure:"endpoint"`
}

// Errors for missing required config parameters.
const (
	ErrNoUsername = "invalid config: missing username"
	ErrNoPassword = "invalid config: missing password"
	ErrNoEndpoint = "invalid config: missing endpoint"
)

func (cfg *Config) Validate() error {
	if cfg.Username == "" {
		return errors.New(ErrNoUsername)
	}
	if cfg.Password == "" {
		return errors.New(ErrNoPassword)
	}
	if cfg.Endpoint == "" {
		return errors.New(ErrNoEndpoint)
	}
	return nil
}
