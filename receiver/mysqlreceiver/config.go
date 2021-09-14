package mysqlreceiver

import (
	"errors"

	"go.opentelemetry.io/collector/receiver/scraperhelper"
	"go.uber.org/multierr"
)

type Config struct {
	scraperhelper.ScraperControllerSettings `mapstructure:",squash"`
	Username                                string `mapstructure:"username"`
	Password                                string `mapstructure:"password"`
	Endpoint                                string `mapstructure:"endpoint"`
}

// Errors for missing required config parameters.
const (
	ErrNoUsername = "invalid config: missing username"
	ErrNoPassword = "invalid config: missing password"
)

func (cfg *Config) Validate() error {
	var errs []error
	if cfg.Username == "" {
		errs = append(errs, errors.New(ErrNoUsername))
	}
	if cfg.Password == "" {
		errs = append(errs, errors.New(ErrNoPassword))
	}
	return multierr.Combine(errs...)
}
