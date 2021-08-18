package postgresqlreceiver

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/multierr"
)

func TestValidate(t *testing.T) {
	testCases := []struct {
		desc     string
		cfg      *Config
		expected error
	}{
		{
			desc: "missing username and password",
			cfg:  &Config{},
			expected: multierr.Combine(
				errors.New(ErrNoUsername),
				errors.New(ErrNoPassword),
			),
		},
		{
			desc: "missing password",
			cfg: &Config{
				Username: "otel",
			},
			expected: multierr.Combine(
				errors.New(ErrNoPassword),
			),
		},
		{
			desc: "missing username",
			cfg: &Config{
				Password: "otel",
			},
			expected: multierr.Combine(
				errors.New(ErrNoUsername),
			),
		},
		{
			desc: "no error",
			cfg: &Config{
				Username: "otel",
				Password: "otel",
			},
			expected: multierr.Combine(),
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			actual := tC.cfg.Validate()
			require.Equal(t, tC.expected, actual)
		})
	}
}
