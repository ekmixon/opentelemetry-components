package elasticsearchreceiver

//go:generate mdatagen metadata.yaml

import (
	"context"
	"time"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config"
	"go.opentelemetry.io/collector/config/confighttp"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/receiver/receiverhelper"
	"go.opentelemetry.io/collector/receiver/scraperhelper"
)

const (
	typeStr = "elasticsearch"
)

// NewFactory creates a factory for elasticsearch receiver.
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
		HTTPClientSettings: confighttp.HTTPClientSettings{
			Endpoint: "localhost:15672",
			Timeout:  10 * time.Second,
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

	ns, err := newElasticSearchScraper(params.Logger, cfg)
	if err != nil {
		return nil, err
	}
	scraper := scraperhelper.NewResourceMetricsScraper(cfg.ID(), ns.scrape, scraperhelper.WithStart(ns.start))

	return scraperhelper.NewScraperControllerReceiver(
		&cfg.ScraperControllerSettings, params.Logger, consumer,
		scraperhelper.AddScraper(scraper),
	)
}
