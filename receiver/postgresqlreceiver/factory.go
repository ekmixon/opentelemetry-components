package postgresqlreceiver

//go:generate mdatagen metadata.yaml

import (
	"context"
	"errors"
	"time"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/receiver/receiverhelper"
	"go.opentelemetry.io/collector/receiver/scraperhelper"
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
	err := validateConfig(cfg)
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

// Errors for missing required config parameters.
const (
	ErrNoUsername = "invalid config: missing username"
	ErrNoPassword = "invalid config: missing password"
	ErrNoEndpoint = "invalid config: missing endpoint"
)

func validateConfig(cfg *Config) error {
	if cfg.Username == "" {
		return errors.New(ErrNoUsername)
	}
	if cfg.Password == "" {
		return errors.New(ErrNoPassword)
	}
	if cfg.Endpoint == "" {
		return errors.New(ErrNoEndpoint)
	}
	return nil
}
