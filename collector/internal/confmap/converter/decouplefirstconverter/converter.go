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

package decouplefirstconverter // import "github.com/open-telemetry/opentelemetry-lambda/collector/internal/confmap/converter/decouplefirstconverter"

import (
	"context"
	"fmt"
	"strings"

	"go.opentelemetry.io/collector/confmap"
	"go.uber.org/zap"
)

const (
	pipelinesKey = "service.pipelines"
	processorsKey = "processors"
)

type converter struct{}

// New returns a confmap.Converter that ensures the decoupleprocessor is placed first in the pipeline.
func New() confmap.Converter {
	return &converter{}
}


func (c converter) Convert(_ context.Context, conf *confmap.Conf) error {
    pipelines, err := conf.Sub(pipelinesKey)
    if err != nil {
        return fmt.Errorf("invalid service.pipelines configuration: %w", err)
    }

	// Iterate pipelines over telemetry types: traces, metrics, logs
    for telemetryType, pipelineVal := range pipelines.ToStringMap() {
        // pipeline, ok := pipelineVal.(map[string]interface{})
        // if !ok {
        //     return fmt.Errorf("invalid pipeline configuration for telemetry type %s", telemetryType)
        // }

		// Get the processors for the pipeline
        processorsKey := fmt.Sprintf("%s.processors", telemetryType)
		processorsSub, err := conf.Sub(processorsKey)
		if err != nil {
			return fmt.Errorf("invalid processors configuration for telemetry type %s: %w", telemetryType, err)
		}
        var processors []interface{}
        if err := processorsSub.Unmarshal(&processors); err != nil {
            return fmt.Errorf("invalid processors configuration for telemetry type %s: %w", telemetryType, err)
            continue
        }

		// If there are processors, check if the first processor is "decouple"
		// and prepend it if not
        if len(processors) > 0 {
            firstProcessor, ok := processors[0].(string)
            if !ok {
                return fmt.Errorf("invalid processor configuration for telemetry type %s", telemetryType)
            }

            if firstProcessor != "decouple" {
                zap.L().Warn("Did not find decoupleprocessor as the first processor in the pipeline. Prepending a decoupleprocessor.")
				// Prepend the decouple processor to the processors if it is not already the first processor
                processors = append([]interface{}{"decouple"}, processors...)
			}

			// Drop all "decouple" processors after the first
			for i, v := range processors[1:] {
				if pstr, ok := v.(string); ok && strings.HasPrefix(pstr, "decouple/") {
					processors = append(processors[:i+1], processors[i+2:]...)
					zap.L().Warn("Decouple processor out of first position. Dropped " + fmt.Sprintf("%d", i))
				}
			}

			// Update the processors configuration
			// pipeline["processors"] = processors
			if err := processorsSub.Marshal(processors); err != nil {
				return fmt.Errorf("failed to update processors configuration for telemetry type %s: %w", telemetryType, err)
			}
        }
    }

    return nil
}
