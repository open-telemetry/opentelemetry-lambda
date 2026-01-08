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

package accountidprocessor

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/confmap"
)

func TestConvert(t *testing.T) {
	tests := []struct {
		name          string
		accountID     string
		input         map[string]any
		expectedProcs map[string]any
		shouldHaveRes bool
	}{
		{
			name:      "empty_account_id",
			accountID: "",
			input: map[string]any{
				"service": map[string]any{
					"pipelines": map[string]any{
						"traces": map[string]any{
							"receivers":  []any{"otlp"},
							"processors": []any{"batch"},
							"exporters":  []any{"otlp"},
						},
					},
				},
			},
			shouldHaveRes: false,
		},
		{
			name:      "no_service",
			accountID: "123456789012",
			input: map[string]any{
				"receivers": map[string]any{},
			},
			shouldHaveRes: false,
		},
		{
			name:      "no_pipelines",
			accountID: "123456789012",
			input: map[string]any{
				"service": map[string]any{},
			},
			shouldHaveRes: false,
		},
		{
			name:      "single_pipeline_with_processors",
			accountID: "123456789012",
			input: map[string]any{
				"service": map[string]any{
					"pipelines": map[string]any{
						"traces": map[string]any{
							"receivers":  []any{"otlp"},
							"processors": []any{"batch"},
							"exporters":  []any{"otlp"},
						},
					},
				},
			},
			expectedProcs: map[string]any{
				"resource/aws-account-id": map[string]any{
					"attributes": []map[string]any{
						{
							"key":    "cloud.account.id",
							"value":  "123456789012",
							"action": "insert",
						},
					},
				},
			},
			shouldHaveRes: true,
		},
		{
			name:      "single_pipeline_no_processors",
			accountID: "123456789012",
			input: map[string]any{
				"service": map[string]any{
					"pipelines": map[string]any{
						"traces": map[string]any{
							"receivers": []any{"otlp"},
							"exporters": []any{"otlp"},
						},
					},
				},
			},
			expectedProcs: map[string]any{
				"resource/aws-account-id": map[string]any{
					"attributes": []map[string]any{
						{
							"key":    "cloud.account.id",
							"value":  "123456789012",
							"action": "insert",
						},
					},
				},
			},
			shouldHaveRes: true,
		},
		{
			name:      "multiple_pipelines",
			accountID: "987654321098",
			input: map[string]any{
				"service": map[string]any{
					"pipelines": map[string]any{
						"traces": map[string]any{
							"receivers":  []any{"otlp"},
							"processors": []any{"batch"},
							"exporters":  []any{"otlp"},
						},
						"logs": map[string]any{
							"receivers":  []any{"otlp"},
							"processors": []any{},
							"exporters":  []any{"otlp"},
						},
						"metrics": map[string]any{
							"receivers": []any{"otlp"},
							"exporters": []any{"prometheus"},
						},
					},
				},
			},
			expectedProcs: map[string]any{
				"resource/aws-account-id": map[string]any{
					"attributes": []map[string]any{
						{
							"key":    "cloud.account.id",
							"value":  "987654321098",
							"action": "insert",
						},
					},
				},
			},
			shouldHaveRes: true,
		},
		{
			name:      "existing_processors",
			accountID: "111111111111",
			input: map[string]any{
				"service": map[string]any{
					"pipelines": map[string]any{
						"traces": map[string]any{
							"receivers":  []any{"otlp"},
							"processors": []any{"batch", "attributes"},
							"exporters":  []any{"otlp"},
						},
					},
				},
			},
			expectedProcs: map[string]any{
				"resource/aws-account-id": map[string]any{
					"attributes": []map[string]any{
						{
							"key":    "cloud.account.id",
							"value":  "111111111111",
							"action": "insert",
						},
					},
				},
			},
			shouldHaveRes: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conf := confmap.NewFromStringMap(tt.input)
			converter := New(tt.accountID)
			err := converter.Convert(context.Background(), conf)
			require.NoError(t, err)

			if !tt.shouldHaveRes {
				// For cases where no resource processor should be added
				if procVal := conf.Get("processors"); procVal != nil {
					procs, ok := procVal.(map[string]any)
					if ok {
						assert.NotContains(t, procs, "resource/aws-account-id")
					}
				}
				return
			}

			// Check that resource processor was added
			procVal := conf.Get("processors")
			require.NotNil(t, procVal)
			procs, ok := procVal.(map[string]any)
			require.True(t, ok, "processors should be a map")

			resourceProc, ok := procs["resource/aws-account-id"]
			require.True(t, ok, "resource/aws-account-id processor should exist")

			// Verify processor configuration
			expectedProc := tt.expectedProcs["resource/aws-account-id"]
			assert.Equal(t, expectedProc, resourceProc)

			// Check that all pipelines have the resource processor prepended
			serviceVal := conf.Get("service")
			require.NotNil(t, serviceVal)
			service, ok := serviceVal.(map[string]any)
			require.True(t, ok)

			pipelinesVal, ok := service["pipelines"]
			require.True(t, ok)
			pipelines, ok := pipelinesVal.(map[string]any)
			require.True(t, ok)

			for pipelineType, pipelineVal := range pipelines {
				pipeline, ok := pipelineVal.(map[string]any)
				require.True(t, ok, "pipeline %s should be a map", pipelineType)

				processorsVal, ok := pipeline["processors"]
				require.True(t, ok, "pipeline %s should have processors", pipelineType)
				processors, ok := processorsVal.([]any)
				require.True(t, ok, "processors should be a slice")

				// First processor should be resource/aws-account-id
				require.Greater(t, len(processors), 0, "pipeline %s should have at least one processor", pipelineType)
				assert.Equal(t, "resource/aws-account-id", processors[0], "first processor in %s should be resource/aws-account-id", pipelineType)
			}
		})
	}
}

func TestConvert_AccountIDValues(t *testing.T) {
	tests := []struct {
		name      string
		accountID string
	}{
		{"12_digits", "123456789012"},
		{"different_account", "999999999999"},
		{"all_zeros", "000000000000"},
		{"sequential", "111111111111"},
		{"leading_zero", "012345678901"},
		{"multiple_leading_zeros", "001234567890"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := map[string]any{
				"service": map[string]any{
					"pipelines": map[string]any{
						"traces": map[string]any{
							"receivers":  []any{"otlp"},
							"processors": []any{},
							"exporters":  []any{"otlp"},
						},
					},
				},
			}

			conf := confmap.NewFromStringMap(input)
			converter := New(tt.accountID)
			err := converter.Convert(context.Background(), conf)
			require.NoError(t, err)

			// Verify the account ID is correctly set
			procVal := conf.Get("processors")
			procs := procVal.(map[string]any)
			resourceProc := procs["resource/aws-account-id"].(map[string]any)
			attributes := resourceProc["attributes"].([]map[string]any)

			require.Equal(t, 1, len(attributes))
			assert.Equal(t, tt.accountID, attributes[0]["value"])
		})
	}
}

func TestNew(t *testing.T) {
	accountID := "123456789012"
	converter := New(accountID)
	assert.NotNil(t, converter)
}
