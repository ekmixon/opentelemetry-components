package httpdreceiver

import (
	"fmt"
	"net"
	"net/url"
	"strings"

	"go.opentelemetry.io/collector/config/confighttp"
	"go.opentelemetry.io/collector/receiver/scraperhelper"
)

type Config struct {
	scraperhelper.ScraperControllerSettings `mapstructure:",squash"`
	confighttp.HTTPClientSettings           `mapstructure:",squash"`
}

// TODO We need a way to validate that this is actually helping.
//        This means:
// 1) We have a specific scenario that fails without it.
// 2) We have a test, that validate that the scenario fails,
//      and that this fixes it.
// 3) Manually validate that it fails without the modification,
//      and that it passes with the modification.

var (
	DefaultScheme   = "http"
	DefaultProtocol = "http://"
	DefaultHost     = "localhost"
	DefaultPort     = ":8081" // 8080 or 8081? we picked 8081 because nginx uses 8080
	DefaultPath     = "/server-status"
	ValidPath       = "/server-status/"
	DefaultQuery    = "auto"
)

func (cfg *Config) Validate() error {
	if cfg.Endpoint != "" {
		cfg.Endpoint = strings.ToLower(cfg.Endpoint)

		if !strings.HasPrefix(cfg.Endpoint, DefaultScheme) { // adds default schema
			cfg.Endpoint = fmt.Sprintf("%s%s", DefaultProtocol, cfg.Endpoint)
		}

		u, err := url.Parse(cfg.Endpoint) // err will never fail here
		if err != nil {
			return fmt.Errorf("invalid endpoint '%s'", cfg.Endpoint)
		}

		if u.Port() == "" { // adds default port
			u.Host = u.Host + DefaultPort
		}

		if !(net.ParseIP(u.Hostname()) != nil || u.Hostname() == DefaultHost) { // validates either IP address or local
			return fmt.Errorf("invalid host: expected IP address or localhost, but got %s", u.Hostname())
		}

		if u.Path == "" { // adds default path
			u.Path = DefaultPath
		} else if !(u.Path == DefaultPath || u.Path == ValidPath) {
			return fmt.Errorf("invalid path: expected '%s' path, but got '%s'", DefaultPath, u.Path)
		}

		if u.RawQuery == "" { // adds default query
			u.RawQuery = DefaultQuery
		} else if u.RawQuery != DefaultQuery {
			return fmt.Errorf("invalid raw query: expected '%s' query, but got '%s'", DefaultQuery, u.RawQuery)
		}

		if u.Fragment != "" { // validates no fragments
			return fmt.Errorf("invalid fragment: expected empty string, but got %s", u.Fragment)
		}

		cfg.Endpoint = u.String()
	} else {
		u := url.URL{
			Scheme:   DefaultScheme,
			Host:     DefaultHost + DefaultPort,
			Path:     DefaultPath,
			RawQuery: DefaultQuery,
		}
		cfg.Endpoint = u.String()
	}
	return nil
}
