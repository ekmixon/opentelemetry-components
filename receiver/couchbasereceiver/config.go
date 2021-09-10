package couchbasereceiver

import (
	"errors"
	"fmt"
	"net/url"
	"strings"

	"go.opentelemetry.io/collector/config/confighttp"
	"go.opentelemetry.io/collector/receiver/scraperhelper"
	"go.uber.org/multierr"
)

type Config struct {
	scraperhelper.ScraperControllerSettings `mapstructure:",squash"`
	confighttp.HTTPClientSettings           `mapstructure:",squash"`
	Username                                string `mapstructure:"username"`
	Password                                string `mapstructure:"password"`
}

var (
	// Defaults for missing parameters
	DefaultProtocol = "http://"
	DefaultHost     = "localhost"
	DefaultPort     = "8091"
	DefaultEndpoint = fmt.Sprintf("%s%s:%s", DefaultProtocol, DefaultHost, DefaultPort)

	// Errors for missing required config parameters.
	ErrNoUsername = "missing required field 'username'"
	ErrNoPassword = "missing required field 'password'"
)

// Validate validates missing and invalid configuration fields.
func (cfg *Config) Validate() error {
	if missingFields := ValidateMissingFields(cfg); missingFields != nil {
		return missingFields
	}
	return cfg.ValidateEndpoint()
}

// ValidateMissingFields validates config values for missing fields.
func ValidateMissingFields(cfg *Config) error {
	var errs []error

	if cfg.Username == "" {
		errs = append(errs, errors.New(ErrNoUsername))
	}
	if cfg.Password == "" {
		errs = append(errs, errors.New(ErrNoPassword))
	}
	return multierr.Combine(errs...)
}

// ValidateEndpoint validates the endpoint by adding sensible default fields.
func (cfg *Config) ValidateEndpoint() error {
	if cfg.Endpoint == "" {
		cfg.Endpoint = DefaultEndpoint
		return nil
	}

	if missingProtocol(cfg.Endpoint) {
		cfg.Endpoint = fmt.Sprintf("%s%s", DefaultProtocol, cfg.Endpoint)
	}

	u, err := url.Parse(cfg.Endpoint)
	if err != nil {
		return fmt.Errorf("invalid endpoint '%s'", cfg.Endpoint)
	}

	if u.Hostname() == "" {
		u.Host = fmt.Sprintf("%s:%s", DefaultHost, DefaultPort)
	}

	if u.Port() == "" {
		u.Host = fmt.Sprintf("%s:%s", u.Host, DefaultPort)
	}

	// the url path/query to scrape metrics gets called in the scraper.
	u.Path = ""
	u.RawQuery = ""

	cfg.Endpoint = u.String()
	return nil
}

// missingProtocol returns true if any http protocol is found, case sensitive.
func missingProtocol(rawUrl string) bool {
	return !strings.HasPrefix(strings.ToLower(rawUrl), "http")
}
