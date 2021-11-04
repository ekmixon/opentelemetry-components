package postgresqlreceiver

//go:generate mdatagen metadata.yaml

import (
	"context"
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
		Host: "localhost",
		Port: 5432,
	}
}

func createMetricsReceiver(
	_ context.Context,
	params component.ReceiverCreateSettings,
	rConf config.Receiver,
	consumer consumer.Metrics,
) (component.MetricsReceiver, error) {
	cfg := rConf.(*Config)

	ns := newPostgreSQLScraper(params.Logger, cfg)
	scraper := scraperhelper.NewResourceMetricsScraper(cfg.ID(), ns.scrape, scraperhelper.WithStart(ns.start),
		scraperhelper.WithShutdown(ns.shutdown))

	return scraperhelper.NewScraperControllerReceiver(
		&cfg.ScraperControllerSettings, params.Logger, consumer,
		scraperhelper.AddScraper(scraper),
	)
}
