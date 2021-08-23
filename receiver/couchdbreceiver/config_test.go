package couchdbreceiver

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/multierr"
)

func TestValidate(t *testing.T) {
	t.Run("error path", func(t *testing.T) {
		cfg := &Config{
			Username: "otelu",
			Password: "otelp",
			Endpoint: "http://endpoint with space",
		}
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
				expected: "http://localhost:5984",
			},
			{
				desc:     "only host(local)",
				rawUrl:   "localhost",
				expected: "http://localhost:5984",
			},
			{
				desc:     "only host",
				rawUrl:   "127.0.0.1",
				expected: "http://127.0.0.1:5984",
			},
			{
				desc:     "host(local) and port",
				rawUrl:   "localhost:5984",
				expected: "http://localhost:5984",
			},
			{
				desc:     "host and port",
				rawUrl:   "127.0.0.1:5984",
				expected: "http://127.0.0.1:5984",
			},
			{
				desc:     "full path",
				rawUrl:   "http://localhost:5984/_node/_local/_stats/couchdb",
				expected: "http://localhost:5984",
			},
			{
				desc:     "full path",
				rawUrl:   "http://127.0.0.1:5984/_node/_local/_stats/couchdb",
				expected: "http://127.0.0.1:5984",
			},
			{
				desc:     "unique host no port",
				rawUrl:   "myAlias.Site",
				expected: "http://myAlias.Site:5984",
			},
			{
				desc:     "unique host with port",
				rawUrl:   "myAlias.Site:1234",
				expected: "http://myAlias.Site:1234",
			},
			{
				desc:     "unique host with port with path",
				rawUrl:   "myAlias.Site:1234/_node/_local/_stats/couchdb",
				expected: "http://myAlias.Site:1234",
			},
			{
				desc:     "only port",
				rawUrl:   ":5984",
				expected: "http://localhost:5984",
			},
			{
				desc:     "limitation: double port",
				rawUrl:   "1234",
				expected: "http://1234:5984",
			},
			{
				desc:     "limitation: invalid ip",
				rawUrl:   "500.0.0.0.1.1",
				expected: "http://500.0.0.0.1.1:5984",
			},
		}
		for _, tC := range testCases {
			t.Run(tC.desc, func(t *testing.T) {
				cfg := &Config{
					Username: "otelu",
					Password: "otelp",
					Endpoint: tC.rawUrl,
				}
				require.NoError(t, cfg.Validate())
				require.Equal(t, tC.expected, cfg.Endpoint)
			})
		}
	})
}

func TestValidateMissingFields(t *testing.T) {
	testCases := []struct {
		desc   string
		cfg    *Config
		actual error
	}{
		{
			desc: "missing username and password",
			cfg:  &Config{},
			actual: multierr.Combine(
				errors.New(ErrNoUsername),
				errors.New(ErrNoPassword),
			),
		},
		{
			desc: "missing password",
			cfg: &Config{
				Username: "otel",
			},
			actual: multierr.Combine(
				errors.New(ErrNoPassword),
			),
		},
		{
			desc: "no error",
			cfg: &Config{
				Username: "otel",
				Password: "otel",
			},
			actual: nil,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			expected := ValidateMissingFields(tC.cfg)
			require.Equal(t, expected, tC.actual)
		})
	}
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
	ports := []string{"", ":5984", ":1234"}
	paths := []string{"", "/server-status", "/nonsense"}
	queries := []string{"", "/?auto", "/?nonsense"}
	endpoints := []string{}
	validEndpoints := map[string]bool{
		"http://127.0.0.1:5984":      true,
		"http://127.0.0.1:1234":      true,
		"http://localhost:5984":      true,
		"http://localhost:1234":      true,
		"http://customhost.com:5984": true,
		"http://customhost.com:1234": true,
		// https
		"https://127.0.0.1:5984":      true,
		"https://127.0.0.1:1234":      true,
		"https://localhost:5984":      true,
		"https://localhost:1234":      true,
		"https://customhost.com:5984": true,
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
		t.Run(endpoint, func(t *testing.T) {
			cfg := &Config{
				Username: "otelu",
				Password: "otelp",
				Endpoint: endpoint,
			}
			err := cfg.Validate()
			require.NoError(t, err)
			_, ok := validEndpoints[cfg.Endpoint]
			require.True(t, ok)
		})
	}
}
