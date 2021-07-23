package postgresqlreceiver

//go:generate mdatagen metadata.yaml

import (
	"context"
	"fmt"
	"net/url"
	"regexp"
	"time"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/receiver/receiverhelper"
	"go.opentelemetry.io/collector/receiver/scraperhelper"
	"go.uber.org/multierr"
)

const (
	typeStr = "postgresql"
)

func NewFactory() component.ReceiverFactory {
	return receiverhelper.NewFactory(
		typeStr,
		createDefaultConfig,
		receiverhelper.WithMetrics(createMetricsReceiver))
}

func createDefaultConfig() config.Receiver {
	return &Config{
		ScraperControllerSettings: scraperhelper.ScraperControllerSettings{
			ReceiverSettings:   config.NewReceiverSettings(config.NewID(typeStr)),
			CollectionInterval: 10 * time.Second,
		},
	}
}

func createMetricsReceiver(
	_ context.Context,
	params component.ReceiverCreateSettings,
	rConf config.Receiver,
	consumer consumer.Metrics,
) (component.MetricsReceiver, error) {
	cfg := rConf.(*Config)
	err := cfg.Validate()
	if err != nil {
		return nil, err
	}

	ns := newPostgreSQLScraper(params.Logger, cfg)
	scraper := scraperhelper.NewResourceMetricsScraper(cfg.ID(), ns.scrape, scraperhelper.WithStart(ns.start),
		scraperhelper.WithShutdown(ns.shutdown))

	return scraperhelper.NewScraperControllerReceiver(
		&cfg.ScraperControllerSettings, params.Logger, consumer,
		scraperhelper.AddScraper(scraper),
	)
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
	} else if _, err := url.Parse(cfg.Endpoint); err != nil {
		errs = append(errs, fmt.Errorf("invalid url specified in field 'endpoint'"))
	} else if !validPort.MatchString(cfg.Endpoint) {
		errs = append(errs, fmt.Errorf("invalid port specified in field 'endpoint'"))
	}
	return multierr.Combine(errs...)
}
