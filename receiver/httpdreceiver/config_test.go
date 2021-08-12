package httpdreceiver

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidate(t *testing.T) {
	t.Run("error path", func(t *testing.T) {
		cfg := NewFactory().CreateDefaultConfig().(*Config)
		cfg.Endpoint = "http://endpoint with space"
		require.Equal(t, errors.New("invalid endpoint 'http://endpoint with space'"), cfg.Validate())
	})

	t.Run("happy path", func(t *testing.T) {
		testCases := []struct {
			desc        string
			rawUrl      string
			expected    error
			expectedUrl string
		}{
			{
				desc:        "default path",
				rawUrl:      "",
				expected:    nil,
				expectedUrl: "http://localhost:8080/server-status?auto",
			},
			{
				desc:        "only host(local)",
				rawUrl:      "localhost",
				expected:    nil,
				expectedUrl: "http://localhost:8080",
			},
			{
				desc:        "only host",
				rawUrl:      "127.0.0.1",
				expected:    nil,
				expectedUrl: "http://127.0.0.1:8080",
			},
			{
				desc:        "only schema(http), host, port, query",
				rawUrl:      "http://127.0.0.1:8080/server-status?auto",
				expected:    nil,
				expectedUrl: "http://127.0.0.1:8080/server-status?auto",
			},
			{
				desc:        "unique host no port",
				rawUrl:      "myAlias.Site",
				expected:    nil,
				expectedUrl: "http://myAlias.Site:8080",
			},
			{
				desc:        "unique host with port",
				rawUrl:      "myAlias.Site:1234",
				expected:    nil,
				expectedUrl: "http://myAlias.Site:1234",
			},
			{
				desc:        "unique port",
				rawUrl:      "8080",
				expected:    nil,
				expectedUrl: "http://8080:8080",
			},
			{
				desc:        "invalid ip",
				rawUrl:      "127.0.0.0.1.1",
				expected:    nil,
				expectedUrl: "http://127.0.0.0.1.1:8080",
			},
			{
				desc:        "multiple ports",
				rawUrl:      "8080:8080:8080",
				expected:    nil,
				expectedUrl: "http://8080:8080:8080",
			},
		}
		for _, tC := range testCases {
			t.Run(tC.desc, func(t *testing.T) {
				cfg := NewFactory().CreateDefaultConfig().(*Config)
				cfg.Endpoint = tC.rawUrl
				require.NoError(t, cfg.Validate())
				require.Equal(t, tC.expectedUrl, cfg.Endpoint)
			})
		}
	})
}

func TestMissingProtocol(t *testing.T) {
	testCases := []struct {
		desc     string
		rawUrl   string
		expected bool
	}{
		{
			desc:     "localhost, no http",
			rawUrl:   "localhost",
			expected: true,
		},
		{
			desc:     "localhost, no http",
			rawUrl:   "127.0.0.1",
			expected: true,
		},
		{
			desc:     "localhost, with http",
			rawUrl:   "http://localhost",
			expected: false,
		},
		{
			desc:     "localhost, with https",
			rawUrl:   "https://localhost",
			expected: false,
		},
		{
			desc:     "localhost, with http",
			rawUrl:   "http://127.0.0.1",
			expected: false,
		},
		{
			desc:     "localhost, with https",
			rawUrl:   "https://127.0.0.1",
			expected: false,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			require.Equal(t, tC.expected, missingProtocol(tC.rawUrl))
		})
	}
}
