package mongodbreceiver

import (
	"errors"
	"time"

	"go.opentelemetry.io/collector/config/confignet"
	"go.opentelemetry.io/collector/config/configtls"
	"go.opentelemetry.io/collector/receiver/scraperhelper"
)

type Config struct {
	scraperhelper.ScraperControllerSettings `mapstructure:",squash"`
	confignet.TCPAddr                       `mapstructure:",squash"`
	configtls.TLSClientSetting              `mapstructure:"tls,omitempty"`
	Username                                string        `mapstructure:"username"`
	Password                                string        `mapstructure:"password"`
	Timeout                                 time.Duration `mapstructure:"timeout"`
}

func (c *Config) Validate() error {
	if c.Username != "" && c.Password == "" {
		return errors.New("user provided without password")
	} else if c.Username == "" && c.Password != "" {
		return errors.New("password provided without user")
	}
	return nil
}
