package mysqlreceiver

import (
	"go.opentelemetry.io/collector/receiver/scraperhelper"
)

type Config struct {
	scraperhelper.ScraperControllerSettings `mapstructure:",squash"`
	Username                                string `mapstructure:"username"`
	Password                                string `mapstructure:"password,omitempty"`
	Database                                string `mapstructure:"database"`
	Endpoint                                string `mapstructure:"endpoint"`
}
