package httpdreceiver

import (
	"errors"
	"fmt"
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
			desc     string
			rawUrl   string
			expected string
		}{
			{
				desc:     "default path",
				rawUrl:   "",
				expected: "http://localhost:8080",
			},
			{
				desc:     "only host(local)",
				rawUrl:   "localhost",
				expected: "http://localhost:8080",
			},
			{
				desc:     "only host",
				rawUrl:   "127.0.0.1",
				expected: "http://127.0.0.1:8080",
			},
			{
				desc:     "host(local) and port",
				rawUrl:   "localhost:8080",
				expected: "http://localhost:8080",
			},
			{
				desc:     "host and port",
				rawUrl:   "127.0.0.1:8080",
				expected: "http://127.0.0.1:8080",
			},
			{
				desc:     "full path",
				rawUrl:   "http://localhost:8080/server-status?auto",
				expected: "http://localhost:8080",
			},
			{
				desc:     "full path",
				rawUrl:   "http://127.0.0.1:8080/server-status?auto",
				expected: "http://127.0.0.1:8080",
			},
			{
				desc:     "unique host no port",
				rawUrl:   "myAlias.Site",
				expected: "http://myAlias.Site:8080",
			},
			{
				desc:     "unique host with port",
				rawUrl:   "myAlias.Site:1234",
				expected: "http://myAlias.Site:1234",
			},
			{
				desc:     "unique host with port with path",
				rawUrl:   "myAlias.Site:1234/server-status?auto",
				expected: "http://myAlias.Site:1234",
			},
			{
				desc:     "only port",
				rawUrl:   ":8080",
				expected: "http://localhost:8080",
			},
			{
				desc:     "limitation: double port",
				rawUrl:   "1234",
				expected: "http://1234:8080",
			},
			{
				desc:     "limitation: invalid ip",
				rawUrl:   "500.0.0.0.1.1",
				expected: "http://500.0.0.0.1.1:8080",
			},
		}
		for _, tC := range testCases {
			t.Run(tC.desc, func(t *testing.T) {
				cfg := NewFactory().CreateDefaultConfig().(*Config)
				cfg.Endpoint = tC.rawUrl
				require.NoError(t, cfg.Validate())
				require.Equal(t, tC.expected, cfg.Endpoint)
			})
		}
	})
}

func TestMissingProtocol(t *testing.T) {
	testCases := []struct {
		desc     string
		proto    string
		expected bool
	}{
		{
			desc:     "http proto",
			proto:    "http://localhost",
			expected: false,
		},
		{
			desc:     "https proto",
			proto:    "https://localhost",
			expected: false,
		},
		{
			desc:     "HTTP caps",
			proto:    "HTTP://localhost",
			expected: false,
		},
		{
			desc:     "everything else",
			proto:    "ht",
			expected: true,
		},
		{
			desc:     "everything else",
			proto:    "localhost",
			expected: true,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			require.Equal(t, tC.expected, missingProtocol(tC.proto))
		})
	}
}

func TestValidateEndpointFormat(t *testing.T) {
	protocols := []string{"", "http://", "https://"}
	hosts := []string{"", "127.0.0.1", "localhost", "customhost.com"}
	ports := []string{"", ":8080", ":1234"}
	paths := []string{"", "/server-status", "/nonsense"}
	queries := []string{"", "/?auto", "/?nonsense"}
	endpoints := []string{}
	validEndpoints := map[string]bool{
		"http://127.0.0.1:8080":      true,
		"http://127.0.0.1:1234":      true,
		"http://localhost:8080":      true,
		"http://localhost:1234":      true,
		"http://customhost.com:8080": true,
		"http://customhost.com:1234": true,
		// https
		"https://127.0.0.1:8080":      true,
		"https://127.0.0.1:1234":      true,
		"https://localhost:8080":      true,
		"https://localhost:1234":      true,
		"https://customhost.com:8080": true,
		"https://customhost.com:1234": true,
	}

	for _, protocol := range protocols {
		for _, host := range hosts {
			for _, port := range ports {
				for _, path := range paths {
					for _, query := range queries {
						endpoint := fmt.Sprintf("%s%s%s%s%s", protocol, host, port, path, query)
						endpoints = append(endpoints, endpoint)
					}
				}
			}
		}
	}

	for _, endpoint := range endpoints {
		cfg := NewFactory().CreateDefaultConfig().(*Config)
		cfg.Endpoint = endpoint
		require.NoError(t, cfg.Validate())
		_, ok := validEndpoints[cfg.Endpoint]
		require.True(t, ok)
	}
}
