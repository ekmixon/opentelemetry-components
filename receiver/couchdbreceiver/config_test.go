package couchdbreceiver

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/multierr"
)

func TestValidate(t *testing.T) {
	testCases := []struct {
		desc   string
		cfg    *Config
		actual error
	}{
		{
			desc: "missing config fields",
			cfg:  &Config{},
			actual: multierr.Combine(
				errors.New(ErrNoUsername),
				errors.New(ErrNoPassword),
				errors.New(ErrNoEndpoint),
			),
		},
		{
			desc: "no error",
			cfg: &Config{
				Username: "otel",
				Password: "otel",
				Endpoint: "localhost:1234",
			},
			actual: nil,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			expected := tC.cfg.Validate()
			require.Equal(t, expected, tC.actual)
		})
	}
}

func TestValidateMissingFields(t *testing.T) {
	testCases := []struct {
		desc   string
		cfg    *Config
		actual error
	}{
		{
			desc: "missing username, password, endpoint",
			cfg:  &Config{},
			actual: multierr.Combine(
				errors.New(ErrNoUsername),
				errors.New(ErrNoPassword),
				errors.New(ErrNoEndpoint),
			),
		},
		{
			desc: "missing password, endpoint",
			cfg: &Config{
				Username: "otel",
			},
			actual: multierr.Combine(
				errors.New(ErrNoPassword),
				errors.New(ErrNoEndpoint),
			),
		},
		{
			desc: "missing endpoint",
			cfg: &Config{
				Username: "otel",
				Password: "otel",
			},
			actual: multierr.Combine(
				errors.New(ErrNoEndpoint),
			),
		},
		{
			desc: "no error",
			cfg: &Config{
				Username: "otel",
				Password: "otel",
				Endpoint: "localhost:1234",
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

func TestCheckInvalidFields(t *testing.T) {
	testCases := []struct {
		desc     string
		endpoint string
		actual   error
	}{
		{
			desc:     "multiple colons",
			endpoint: "localhost:1234:1234",
			actual:   errors.New(ErrInvalidEndpoint),
		},
		{
			desc:     "no host",
			endpoint: ":1234",
			actual:   errors.New(ErrNoHost),
		},
		{
			desc:     "no port",
			endpoint: "localhost:",
			actual:   errors.New(ErrNoPort),
		},
		{
			desc:     "invalid port",
			endpoint: "localhost:acbd",
			actual:   errors.New(ErrInvalidPort),
		},
		{
			desc:     "no error",
			endpoint: "localhost:1234",
			actual:   nil,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			expected := ValidateEndpoint(tC.endpoint)
			require.Equal(t, expected, tC.actual)
		})
	}
}

func TestMakeClientEndpoint(t *testing.T) {
	cfg := Config{
		Username: "otelu",
		Password: "otelp",
		Endpoint: "localhost:5984",
		Nodename: "_local",
	}
	expected := cfg.MakeClientEndpoint()
	actual := "http://localhost:5984/_node/_local/_stats/couchdb"
	require.Equal(t, expected, actual)
}
