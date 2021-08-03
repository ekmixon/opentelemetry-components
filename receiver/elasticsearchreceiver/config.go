package elasticsearchreceiver

import (
	"fmt"
	"net/url"

	"go.opentelemetry.io/collector/config/confighttp"
	"go.opentelemetry.io/collector/receiver/scraperhelper"
	"go.uber.org/multierr"
)

type Config struct {
	scraperhelper.ScraperControllerSettings `mapstructure:",squash"`
	confighttp.HTTPClientSettings           `mapstructure:",squash"`

	Password string `mapstructure:"password"`
	Username string `mapstructure:"username"`
}

func (cfg *Config) Validate() error {
	var errs []error
	if cfg.Username == "" && cfg.Password != "" {
		errs = append(errs, fmt.Errorf("'password' specified but not 'username'"))
	} else if cfg.Password == "" && cfg.Username != "" {
		errs = append(errs, fmt.Errorf("'username' specified but not 'password'"))
	}
	if cfg.Endpoint == "" {
		errs = append(errs, fmt.Errorf("missing required field 'endpoint'"))
	} else if _, err := url.Parse(cfg.Endpoint); err != nil {
		errs = append(errs, fmt.Errorf("invalid url specified in field 'endpoint'"))
	}
	return multierr.Combine(errs...)
}
