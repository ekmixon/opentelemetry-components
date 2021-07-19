package mongodbreceiver

import (
	"errors"
	"time"

	"go.opentelemetry.io/collector/config/confignet"
	"go.opentelemetry.io/collector/receiver/scraperhelper"
)

type Config struct {
	scraperhelper.ScraperControllerSettings `mapstructure:",squash"`
	confignet.TCPAddr                       `mapstructure:",squash"`
	User                                    *string       `mapstructure:"user"`
	Password                                *string       `mapstructure:"password"`
	Timeout                                 time.Duration `mapstructure:"timeout"`
}

func (c *Config) Validate() error {
	if c.User != nil && c == nil {
		return errors.New("user provided without password")
	} else if c.User == nil && c.Password != nil {
		return errors.New("password provided without user")
	}

	return nil
}
