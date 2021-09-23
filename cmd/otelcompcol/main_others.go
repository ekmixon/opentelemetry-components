//go:build !windows
// +build !windows

package main

import "go.opentelemetry.io/collector/service"

func run(settings service.CollectorSettings) error {
	return runInteractive(settings)
}
