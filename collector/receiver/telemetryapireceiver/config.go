// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package telemetryapireceiver // import "github.com/open-telemetry/opentelemetry-lambda/collector/receiver/telemetryapireceiver"

import (
	"fmt"
)

// Config defines the configuration for the various elements of the receiver agent.
type Config struct {
	extensionID string
	Port        int      `mapstructure:"port"`
	Types       []string `mapstructure:"types"`
}

// Validate validates the configuration by checking for missing or invalid fields
func (cfg *Config) Validate() error {
	for _, t := range cfg.Types {
		if t != platform && t != function && t != extension {
			return fmt.Errorf("unknown extension type: %s", t)
		}
	}
	return nil
}
