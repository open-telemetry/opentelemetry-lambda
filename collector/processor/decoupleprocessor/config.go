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

package decoupleprocessor // import "github.com/open-telemetry/opentelemetry-lambda/collector/processor/decoupleprocessor"
import "errors"

// Config defines the configuration for the various elements of the processor.
type Config struct {
	MaxQueueSize uint32 `mapstructure:"max_queue_size"`
}

var invalidMaxQueueSizeError = errors.New("max_queue_size must be greater than 0")

// Validate validates the configuration by checking for missing or invalid fields
func (cfg *Config) Validate() error {
	if cfg.MaxQueueSize == 0 {
		return invalidMaxQueueSizeError
	}
	return nil
}
