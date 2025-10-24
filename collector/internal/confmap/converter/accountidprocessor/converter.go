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
	"fmt"

	"go.opentelemetry.io/collector/confmap"
)

const (
	serviceKey       = "service"
	pipelinesKey     = "pipelines"
	processorsKey    = "processors"
	resourceProc     = "resource/aws-account-id"
	accountIDAttrKey = "cloud.account.id"
)

type converter struct {
	accountID string
}

// New returns a confmap.Converter that injects cloud.account.id into all pipelines
func New(accountID string) confmap.Converter {
	return &converter{accountID: accountID}
}

func (c converter) Convert(_ context.Context, conf *confmap.Conf) error {
	if c.accountID == "" {
		return nil // Skip if no account ID
	}

	// Navigate to service.pipelines
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

	updates := make(map[string]interface{})

	// For each pipeline, add resource processor to beginning
	for telemetryType, pipelineVal := range pipelines {
		pipeline, ok := pipelineVal.(map[string]interface{})
		if !ok {
			continue
		}

		processorsVal, _ := pipeline[processorsKey]
		processors, ok := processorsVal.([]interface{})
		if !ok {
			processors = []interface{}{}
		}

		// Prepend resource/aws-account-id processor
		processors = append([]interface{}{resourceProc}, processors...)
		updates[fmt.Sprintf("%s::%s::%s::%s", serviceKey, pipelinesKey, telemetryType, processorsKey)] = processors
	}

	// Configure the resource processor with cloud.account.id attribute
	updates[fmt.Sprintf("processors::%s::attributes", resourceProc)] = []map[string]interface{}{
		{
			"key":    accountIDAttrKey,
			"value":  c.accountID,
			"action": "insert",
		},
	}

	// Apply all updates
	if len(updates) > 0 {
		if err := conf.Merge(confmap.NewFromStringMap(updates)); err != nil {
			return err
		}
	}

	return nil
}
