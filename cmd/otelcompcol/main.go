// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"fmt"
	"log"
	"os"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/service"
)

func main() {
	factories, err := components()
	if err != nil {
		log.Fatalf("Failed to build default components: %v", err)
	}

	bi := component.BuildInfo{
		Command:     os.Args[0],
		Description: "observIQ's opentelemetry-collector distribution",
	}

	settings := service.CollectorSettings{Factories: factories, BuildInfo: bi}

	if err := run(settings); err != nil {
		log.Fatal(err)
	}

}

func runInteractive(settings service.CollectorSettings) error {
	cmd := service.NewCommand(settings)

	err := cmd.Execute()
	if err != nil {
		return fmt.Errorf("application run finished with error: %w", err)
	}

	return nil
}
