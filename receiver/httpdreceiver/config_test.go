package httpdreceiver

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/config/confighttp"
)

func TestDefaultConfig(t *testing.T) {

}

func TestValidate(t *testing.T) {

	needsFixup := []string{
		"127.0.0.1/server-status?auto",
		"localhost:8081/server-status?auto",
	}
	testCases := []string{
		"http://127.0.0.1/server-status?auto",
		"https://127.0.0.1/server-status?auto",
		"HTTP://127.0.0.1/server-status?auto",
		"HTTPS://127.0.0.1/server-status?auto",
		"http://localhost:8081/server-status?auto",
		"https://localhost:8081/server-status?auto",
		"HTTP://localhost:8081/server-status?auto",
		"HTTPS://localhost:8081/server-status?auto",
	}

	for _, tc := range needsFixup {
		t.Run(tc, func(t *testing.T) {
			cfg := NewFactory().CreateDefaultConfig().(*Config)
			require.Error(t, cfg.Validate())
			cfg.Fixup()
			require.NoError(t, cfg.Validate())
		})
	}

	for _, tc := range testCases {
		t.Run(tc, func(t *testing.T) {
			cfg := NewFactory().CreateDefaultConfig().(*Config)
			cfg.Endpoint = tc
			require.NoError(t, cfg.Validate())
		})
	}
}
