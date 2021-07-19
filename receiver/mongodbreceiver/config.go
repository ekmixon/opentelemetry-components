
package mongodbreceiver

import (
	"time"

	"go.opentelemetry.io/collector/config/confignet"
	"go.opentelemetry.io/collector/receiver/scraperhelper"
)

type Config struct {
	scraperhelper.ScraperControllerSettings `mapstructure:",squash"`
	confignet.TCPAddr                       `mapstructure:",squash"`
	User                                    *string       `mapstructure:"user"`
	Password                                *string       `mapstructure:"password"`
	Timeout                                 time.Duration `mapstructure:"timeout"`
}
