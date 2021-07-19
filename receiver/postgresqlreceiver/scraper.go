package postgresqlreceiver

import (
	"context"
	"errors"
	"sync"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/model/pdata"
	"go.uber.org/zap"
)

type postgreSQLScraper struct {
	client   client
	stopOnce sync.Once

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
		databaseName: m.config.Database,
		endpoint:     m.config.Endpoint,
	})
	if err != nil {
		return err
	}
	m.client = client

	return nil
}

// shutdown closes open connections.
func (p *postgreSQLScraper) shutdown(context.Context) error {
	var err error
	p.stopOnce.Do(func() {
		err = p.client.Close()
	})
	return err
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
