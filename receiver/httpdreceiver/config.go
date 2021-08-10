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

func (cfg *Config) Validate() error {
	if cfg.Endpoint != "" {

		// TODO We need a way to validate that this is actually helping.
		//        This means:
		// 1) We have a specific scenario that fails without it.
		// 2) We have a test, that validate that the scenario fails,
		//      and that this fixes it.
		// 3) Manually validate that it fails without the modification,
		//      and that it passes with the modification.
		if !strings.HasPrefix(strings.ToLower(cfg.Endpoint), "http") {
			cfg.Endpoint = fmt.Sprintf("http://%s", cfg.Endpoint)
		}

		_, err := url.Parse(cfg.Endpoint)
		if err != nil {
			return fmt.Errorf("invalid url specified in field 'endpoint'")
		}
	}
	return nil
}
