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

// The accountidprocessor implements the Converter for mutating Collector
// configurations to automatically inject the cloud.account.id attribute
// via a resource processor into all pipelines.
package accountidprocessor

import (
	"context"

	"go.opentelemetry.io/collector/confmap"
	semconv "go.opentelemetry.io/otel/semconv/v1.27.0"
)

const (
	serviceKey    = "service"
	pipelinesKey  = "pipelines"
	processorsKey = "processors"
	resourceProc  = "resource/aws-account-id"
)

type converter struct {
	accountID string
}

// New returns a confmap.Converter that injects cloud.account.id into all pipelines
func New(accountID string) confmap.Converter {
	return &converter{accountID: accountID}
}

func (c *converter) Convert(_ context.Context, conf *confmap.Conf) error {
	if c.accountID == "" {
		return nil
	}

	service, ok := conf.Get(serviceKey).(map[string]any)
	if !ok {
		return nil
	}

	pipelines, ok := service[pipelinesKey].(map[string]any)
	if !ok {
		return nil
	}

	updates := make(map[string]any)

	for pipelineName, pipelineVal := range pipelines {
		pipeline, ok := pipelineVal.(map[string]any)
		if !ok {
			continue
		}

		processorsVal, _ := pipeline[processorsKey]
		processors, _ := processorsVal.([]any)

		// Idempotency: skip if already prepended
		if len(processors) > 0 && processors[0] == resourceProc {
			continue
		}

		processors = append([]any{resourceProc}, processors...)
		updates[serviceKey+"::"+pipelinesKey+"::"+pipelineName+"::"+processorsKey] = processors
	}

	// Add the resource processor definition
	updates["processors::"+resourceProc+"::attributes"] = []map[string]any{
		{
			"key":    string(semconv.CloudAccountIDKey),
			"value":  c.accountID,
			"action": "insert",
		},
	}

	return conf.Merge(confmap.NewFromStringMap(updates))
}
