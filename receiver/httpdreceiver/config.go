package httpdreceiver

import (
	"fmt"
	"net/url"
	"strings"

	"go.opentelemetry.io/collector/config/confighttp"
	"go.opentelemetry.io/collector/receiver/scraperhelper"
)

type Config struct {
	scraperhelper.ScraperControllerSettings `mapstructure:",squash"`
	confighttp.HTTPClientSettings           `mapstructure:",squash"`
}

var (
	DefaultProtocol = "http://"
	DefaultHost     = "localhost"
	DefaultPort     = "8080"
	DefaultEndpoint = fmt.Sprintf("%s%s:%s", DefaultProtocol, DefaultHost, DefaultPort)
)

func (cfg *Config) Validate() error {
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
