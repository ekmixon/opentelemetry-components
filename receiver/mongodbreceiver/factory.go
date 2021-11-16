package mongodbreceiver

//go:generate mdatagen metadata.yaml

import (
	"context"
	"time"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config"
	"go.opentelemetry.io/collector/config/confignet"
	"go.opentelemetry.io/collector/config/configtls"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/receiver/receiverhelper"
	"go.opentelemetry.io/collector/receiver/scraperhelper"
)

const (
	typeStr = "mongodb"
)

// NewFactory creates a factory for mongodb receiver.
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
			CollectionInterval: 60 * time.Second,
		},
		Timeout: 10 * time.Second,
		TCPAddr: confignet.TCPAddr{
			Endpoint: "localhost:27017",
		},
		TLSClientSetting: configtls.TLSClientSetting{
			Insecure:           true,
			InsecureSkipVerify: true,
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

	ns := newMongodbScraper(params.Logger, cfg)
	scraper := scraperhelper.NewResourceMetricsScraper(cfg.ID(), ns.scrape, scraperhelper.WithStart(ns.start), scraperhelper.WithShutdown(ns.shutdown))

	return scraperhelper.NewScraperControllerReceiver(
		&cfg.ScraperControllerSettings, params.Logger, consumer,
		scraperhelper.AddScraper(scraper),
	)
}
