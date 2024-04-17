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

// The decoupleafterbatchconverter implements the Converter for mutating Collector
// configurations to ensure the decouple processor is placed after the batch processor.
// This is logically implemented by appending the decouple processor to the end of
// processor chains where a batch processor is found unless another decouple processor
// was seen.
package decoupleafterbatchconverter

import (
	"context"
	"fmt"
	"strings"

	"go.opentelemetry.io/collector/confmap"
)

const (
	serviceKey        = "service"
	pipelinesKey      = "pipelines"
	processorsKey     = "processors"
	batchProcessor    = "batch"
	decoupleProcessor = "decouple"
)

type converter struct{}

// New returns a confmap.Converter that ensures the decoupleprocessor is placed first in the pipeline.
func New() confmap.Converter {
	return &converter{}
}

func (c converter) Convert(_ context.Context, conf *confmap.Conf) error {
	serviceVal := conf.Get(serviceKey)
	service, ok := serviceVal.(map[string]interface{})
	if !ok {
		return nil
	}

	pipelinesVal, ok := service[pipelinesKey]
	if !ok {
		return nil
	}

	pipelines, ok := pipelinesVal.(map[string]interface{})
	if !ok {
		return nil
	}

	// accumulates updates over the pipelines and applies them
	// once all pipeline configs are processed
	updates := make(map[string]interface{})
	for telemetryType, pipelineVal := range pipelines {
		pipeline, ok := pipelineVal.(map[string]interface{})
		if !ok {
			continue
		}

		processorsVal, ok := pipeline[processorsKey]
		if !ok {
			continue
		}

		processors, ok := processorsVal.([]interface{})
		if !ok {
			continue
		}

		// accumulate config updates
		if shouldAppendDecouple(processors) {
			processors = append(processors, decoupleProcessor)
			updates[fmt.Sprintf("%s::%s::%s::%s", serviceKey, pipelinesKey, telemetryType, processorsKey)] = processors
			break
		}

	}

	// apply all updates
	if len(updates) > 0 {
		if err := conf.Merge(confmap.NewFromStringMap(updates)); err != nil {
			return err
		}
	}

	return nil
}

// The shouldAppendDecouple is the filter predicate for the Convert function action. It tells whether
// (bool) there was a decouple processor after the last
// batch processor, which Convert uses to decide whether to append the decouple processor.
func shouldAppendDecouple(processors []interface{}) bool {
	var shouldAppendDecouple bool
	for _, processorVal := range processors {
		processor, ok := processorVal.(string)
		if !ok {
			continue
		}
		processorBaseName := strings.Split(processor, "/")[0]
		if processorBaseName == batchProcessor {
			shouldAppendDecouple = true
		} else if processorBaseName == decoupleProcessor {
			shouldAppendDecouple = false
		}
	}
	return shouldAppendDecouple
}
