package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestComponents(t *testing.T) {
	factories, err := components()

	require.NoError(t, err)

	// Receivers
	require.NotNil(t, factories.Receivers["hostmetrics"])
	require.NotNil(t, factories.Receivers["otlp"])

	// Processors
	require.NotNil(t, factories.Processors["attributes"])
	require.NotNil(t, factories.Processors["filter"])
	require.NotNil(t, factories.Processors["normalizesums"])
	require.NotNil(t, factories.Processors["resource"])
	require.NotNil(t, factories.Processors["resourcedetection"])

	// Exporters
	require.NotNil(t, factories.Exporters["file"])
	require.NotNil(t, factories.Exporters["googlecloud"])
	require.NotNil(t, factories.Exporters["logging"])
	require.NotNil(t, factories.Exporters["observiq"])
	require.NotNil(t, factories.Exporters["otlp"])
	require.NotNil(t, factories.Exporters["otlphttp"])

	// Extensions
	require.NotNil(t, factories.Extensions["pprof"])
	require.NotNil(t, factories.Extensions["zpages"])
}
