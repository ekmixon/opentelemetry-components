package couchdbreceiver

import (
	"errors"
	"fmt"
	"net/url"
	"strconv"
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
	Nodename                                string `mapstructure:"nodename"`
	Endpoint                                string `mapstructure:"endpoint"`
}

const (
	// Errors for missing required config parameters.
	ErrNoUsername = "missing required field 'username'"
	ErrNoPassword = "missing required field 'password'"
	ErrNoEndpoint = "missing required field 'endpoint'"
	ErrNoHost     = "missing host in 'host:port' in required field 'endpoint'"
	ErrNoPort     = "missing port in 'host:port' in required field 'endpoint'"

	// Errors for invalid required config parameters.
	ErrInvalidEndpoint = "invalid 'host':'port' in required field 'endpoint'"
	ErrInvalidHost     = "invalid host in required field 'endpoint'"
	ErrInvalidPort     = "invalid port in required field 'endpoint'"
)

// Validate validates missing and invalid configuration fields.
func (cfg *Config) Validate() error {
	if missingFields := ValidateMissingFields(cfg); missingFields != nil {
		return missingFields
	}
	return ValidateEndpoint(cfg.Endpoint)
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
	if cfg.Endpoint == "" {
		errs = append(errs, errors.New(ErrNoEndpoint))
	}
	return multierr.Combine(errs...)
}

// ValidateEndpoint validates endpoint fields.
func ValidateEndpoint(endpoint string) error {
	hostPort := strings.Split(endpoint, ":")
	if len(hostPort) != 2 {
		return errors.New(ErrInvalidEndpoint)
	}
	if len(hostPort[0]) == 0 {
		return errors.New(ErrNoHost)
	}
	if len(hostPort[1]) == 0 {
		return errors.New(ErrNoPort)
	}

	if _, err := url.Parse(hostPort[0]); err != nil {
		return errors.New(ErrInvalidHost)
	}

	if _, err := strconv.Atoi(hostPort[1]); err != nil {
		return errors.New(ErrInvalidPort)
	}
	return nil
}

// MakeClientEndpoint makes the client endpoint using config parameters.
func (cfg *Config) MakeClientEndpoint() string {
	endpoint := strings.Split(cfg.Endpoint, ":")
	return fmt.Sprintf("http://%s:%s@%s:%s/_node/%s/_stats/couchdb", cfg.Username, cfg.Password, endpoint[0], endpoint[1], cfg.Nodename)
}
