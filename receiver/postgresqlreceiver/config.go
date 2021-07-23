package postgresqlreceiver

import (
	"fmt"
	"net/url"
	"regexp"

	"go.opentelemetry.io/collector/receiver/scraperhelper"
	"go.uber.org/multierr"
)

type Config struct {
	scraperhelper.ScraperControllerSettings `mapstructure:",squash"`
	Username                                string `mapstructure:"username"`
	Password                                string `mapstructure:"password"`
	Database                                string `mapstructure:"database"`
	Endpoint                                string `mapstructure:"endpoint"`
}

func (cfg *Config) Validate() error {
	var errs []error
	var validPort = regexp.MustCompile(`(:\d+)`)

	if cfg.Username == "" {
		errs = append(errs, fmt.Errorf("missing required field 'username'"))
	}
	if cfg.Password == "" {
		errs = append(errs, fmt.Errorf("missing required field 'password'"))
	}
	if cfg.Database == "" {
		errs = append(errs, fmt.Errorf("missing required field 'database'"))
	}
	if cfg.Endpoint == "" {
		errs = append(errs, fmt.Errorf("missing required field 'endpoint'"))
	} else if !validPort.MatchString(cfg.Endpoint) {
		errs = append(errs, fmt.Errorf("invalid port specified in field 'endpoint'"))
	} else if _, err := url.Parse(cfg.Endpoint); err != nil {
		errs = append(errs, fmt.Errorf("invalid url specified in field 'endpoint'"))
	}
	return multierr.Combine(errs...)
}
