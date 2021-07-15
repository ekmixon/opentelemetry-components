package postgresqlreceiver

import (
	"go.opentelemetry.io/collector/receiver/scraperhelper"
)

type Config struct {
	scraperhelper.ScraperControllerSettings `mapstructure:",squash"`
	Username                                string `mapstructure:"username"`
	Password                                string `mapstructure:"password"`
	DatabaseName                            string `mapstructure:"database_name"`
	Endpoint                                string `mapstructure:"endpoint"`
}
