package mysqlreceiver

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
			desc: "missing username",
			cfg: &Config{
				Password: "otel",
			},
			actual: multierr.Combine(
				errors.New(ErrNoUsername),
			),
		},
		{
			desc: "no error",
			cfg: &Config{
				Username: "otel",
				Password: "otel",
			},
			actual: multierr.Combine(),
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			expected := tC.cfg.Validate()
			require.Equal(t, expected, tC.actual)
		})
	}
}
