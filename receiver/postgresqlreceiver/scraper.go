package postgresqlreceiver

import (
	"context"
	"errors"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/model/pdata"
	"go.uber.org/zap"
)

type postgreSQLScraper struct {
	client client

	logger *zap.Logger
	config *Config
}

func newPostgreSQLScraper(
	logger *zap.Logger,
	config *Config,
) *postgreSQLScraper {
	return &postgreSQLScraper{
		logger: logger,
		config: config,
	}
}

// start starts the scraper by initializing the db client connection.
func (m *postgreSQLScraper) start(_ context.Context, host component.Host) error {
	client, err := newPostgreSQLClient(postgreSQLConfig{
		username:     m.config.Username,
		password:     m.config.Password,
		databaseName: m.config.DatabaseName,
		endpoint:     m.config.Endpoint,
	})
	if err != nil {
		return err
	}
	m.client = client

	return nil
}

// shutdown closes open connections.
func (m *postgreSQLScraper) shutdown(context.Context) error {
	if !m.client.Closed() {
		m.logger.Info("gracefully shutdown")
		return m.client.Close()
	}
	return nil
}

// scrape scrapes the mysql db metric stats, transforms them and labels them into a metric slices.
func (p *postgreSQLScraper) scrape(context.Context) (pdata.ResourceMetricsSlice, error) {

	if p.client == nil {
		return pdata.ResourceMetricsSlice{}, errors.New("failed to connect to http client")
	}

	// metric initialization
	rms := pdata.NewResourceMetricsSlice()
	ilm := rms.AppendEmpty().InstrumentationLibraryMetrics().AppendEmpty()
	ilm.InstrumentationLibrary().SetName("otel/mysql")
	// now := pdata.TimestampFromTime(time.Now())

	p.logger.Error("Hello world")

	return rms, nil
}
