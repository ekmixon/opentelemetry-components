package elasticsearchreceiver

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

	Password string `mapstructure:"password"`
	Username string `mapstructure:"username"`
}

var (
	DefaultProtocol = "http://"
	DefaultHost     = "localhost"
	DefaultPort     = "9200"
	DefaultEndpoint = "http://localhost:9200"
)

func (cfg *Config) Validate() error {
	if err := invalidCredentials(cfg.Username, cfg.Password); err != nil {
		return err
	}

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

	cfg.Endpoint = u.String()
	return nil
}

// missingProtocol returns true if any http protocol is found, case sensitive.
func missingProtocol(rawUrl string) bool {
	return !strings.HasPrefix(strings.ToLower(rawUrl), "http")
}

// invalidCredentials returns true if only one username or password is not empty.
func invalidCredentials(username, password string) error {
	if username == "" && password != "" {
		return fmt.Errorf("'password' specified but not 'username'")
	}

	if password == "" && username != "" {
		return fmt.Errorf("'username' specified but not 'password'")
	}
	return nil
}
