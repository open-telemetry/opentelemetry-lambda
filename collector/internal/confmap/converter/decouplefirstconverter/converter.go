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
)

const (
	pipelinesKey  = "service::pipelines"
	processorsKey = "processors"
)

type converter struct{}

// New returns a confmap.Converter that ensures the decoupleprocessor is placed first in the pipeline.
func New() confmap.Converter {
	return &converter{}
}

type arr []interface{}

func (a arr) moveLast(x int) error {
	if x < 0 || x >= len(a) {
		return fmt.Errorf("index out of bounds: %d", x)
	}
	return nil
}

func (a arr) move2ndLast(x int) error {
	// if index out of bounds then return error
	return nil
}

func (a arr) drop(x int) error {
	return nil
}

func (c converter) convertProcessors(processors []interface{}) []interface{} {
	// Drop occurrences of "batch" and "decouple".
	// This ignores edge cases and if user has configured processors earlier in the pipeline then
	// implicit handling is that the config is gone. This only effects batch.
	if len(processors) == 0 {
		return []any{
			string("batch"),
			string("decouple"),
		}
	}
	result := make([]interface{}, 0, len(processors))
	result = addBatchAndDecouple(
		dropBatchAndDecouple(
			processors,
		),
	)

	return result
}

func (c converter) Convert(_ context.Context, conf *confmap.Conf) error {
	pipelines, err := conf.Sub("service::pipelines") // conf.Sub("services::pipelines")
	if err != nil {
		return fmt.Errorf("invalid service.pipelines configuration: %w", err)
	}

	// Iterate pipelines over telemetry types: traces, metrics, logs
	for telemetryType, pipelineVal := range pipelines.ToStringMap() {
		println("telemetryType: %s", telemetryType)
		// Get the processors for the pipeline
		// processorsKey := fmt.Sprintf("%s::processors", telemetryType)
		// extract the processors from the pipeline
		pipeline, ok := pipelineVal.(map[string]interface{})
		if !ok {
			return fmt.Errorf("invalid processors configuration for telemetry type %s", telemetryType)
		}
		processors, ok := pipeline["processors"].([]interface{})
		if !ok {
			return fmt.Errorf("invalid processors configuration for telemetry type %s", telemetryType)
		}
		resultProcessors := c.convertProcessors(processors)
		println("extracted ", len(processors), " processors, from key: ", telemetryType)
		println("converted to ", len(resultProcessors), " processors")
		conf.Merge(confmap.NewFromStringMap(map[string]interface{}{
			"service::pipelines::" + telemetryType + "::processors": resultProcessors,
		}))
	}

	return nil
}

// Drop all occurrences of "batch" and "decouple" processors from the pipeline.
func dropBatchAndDecouple(processors []interface{}) []interface{} {
	// Drop occurrences of "batch" and "decouple".
	// This ignores edge cases and if user has configured processors earlier in the pipeline then
	// implicit handling is that the config is gone. This only effects batch.
	if len(processors) == 0 {
		return make([]interface{}, 0)
	}
	for i, v := range processors {
		if pstr, ok := v.(string); ok && (strings.HasPrefix(pstr, "batch") || strings.HasPrefix(pstr, "decouple")) {
			if i < len(processors) {
				processors = append(processors[:i], processors[i+1:]...)
			}
		}
	}
	return processors
}

// Add the "batch" and "decouple" processors to the pipeline. This is a simplistic implementation
// that assumes the default processors aren't already in the pipeline, which is valid when
// the pipeline filter to drop the processors is applied first.
func addBatchAndDecouple(processors []interface{}) []interface{} {
	suffix := []interface{}{"batch", "decouple"}
	if processors == nil || len(processors) == 0 {
		return suffix
	}
	result := make([]interface{}, 0, len(processors)+len(suffix))
	result = append(result, processors...)
	result = append(result, suffix...)

	return result
}
