// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package decoupleafterbatchconverter

import (
	"context"
	"testing"

	"go.opentelemetry.io/collector/confmap"

	"github.com/google/go-cmp/cmp"
)

func TestConvert(t *testing.T) {
	// Since this really tests differences in input, it's easier to read cases
	// without the repeated definition of other fields in the config.
	baseConf := func(input []interface{}) *confmap.Conf {
		return confmap.NewFromStringMap(map[string]interface{}{
			"service": map[string]interface{}{
				"pipelines": map[string]interface{}{
					"traces": map[string]interface{}{
						"processors": input,
					},
				},
			},
		})
	}

	testCases := []struct {
		name     string
		input    *confmap.Conf
		expected *confmap.Conf
		err      error
	}{
		// This test is first, because it illustrates the difference in making the rule that when
		// batch is present the converter appends decouple processor to the end of chain versus
		// the approach of this code which is to do this only when the last instance of batch
		// is not followed by decouple processor.
		{
			name:     "batch then decouple in middle of chain",
			input:    baseConf([]interface{}{"processor1", "batch", "decouple", "processor2"}),
			expected: baseConf([]interface{}{"processor1", "batch", "decouple", "processor2"}),
		},
		{
			name:     "no service",
			input:    confmap.New(),
			expected: confmap.New(),
		},
		{
			name: "no pipelines",
			input: confmap.NewFromStringMap(
				map[string]interface{}{
					"service": map[string]interface{}{
						"extensions": map[string]interface{}{},
					},
				},
			),
			expected: confmap.NewFromStringMap(
				map[string]interface{}{
					"service": map[string]interface{}{
						"extensions": map[string]interface{}{},
					},
				},
			),
		},
		{
			name: "no processors in chain",
			input: confmap.NewFromStringMap(
				map[string]interface{}{
					"service": map[string]interface{}{
						"extensions": map[string]interface{}{},
						"pipelines": map[string]interface{}{
							"traces": map[string]interface{}{},
						},
					},
				},
			),
			expected: confmap.NewFromStringMap(map[string]interface{}{
				"service": map[string]interface{}{
					"extensions": map[string]interface{}{},
					"pipelines": map[string]interface{}{
						"traces": map[string]interface{}{},
					},
				},
			},
			),
		},
		{
			name:     "batch processor in singleton chain",
			input:    baseConf([]interface{}{"batch"}),
			expected: baseConf([]interface{}{"batch", "decouple"}),
		},
		{
			name:     "batch processor present twice",
			input:    baseConf([]interface{}{"batch", "processor1", "batch"}),
			expected: baseConf([]interface{}{"batch", "processor1", "batch", "decouple"}),
		},

		{
			name:     "batch processor not present",
			input:    baseConf([]interface{}{"processor1", "processor2"}),
			expected: baseConf([]interface{}{"processor1", "processor2"}),
		},
		{
			name:     "batch sandwiched between input no decouple",
			input:    baseConf([]interface{}{"processor1", "batch", "processor2"}),
			expected: baseConf([]interface{}{"processor1", "batch", "processor2", "decouple"}),
		},

		{
			name:     "batch and decouple input already present in correct position",
			input:    baseConf([]interface{}{"processor1", "batch", "processor2", "decouple"}),
			expected: baseConf([]interface{}{"processor1", "batch", "processor2", "decouple"}),
		},
		{
			name:     "decouple and batch",
			input:    baseConf([]interface{}{"decouple", "batch"}),
			expected: baseConf([]interface{}{"decouple", "batch", "decouple"}),
		},
		{
			name:     "decouple then batch mixed with others in the pipelinefirst then batch somewhere",
			input:    baseConf([]interface{}{"processor1", "decouple", "processor2", "batch", "processor3"}),
			expected: baseConf([]interface{}{"processor1", "decouple", "processor2", "batch", "processor3", "decouple"}),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			conf := tc.input
			expected := tc.expected

			c := New()
			err := c.Convert(context.Background(), conf)
			if err != tc.err {
				t.Errorf("unexpected error converting: %v", err)
			}
			if diff := cmp.Diff(expected.ToStringMap(), conf.ToStringMap()); diff != "" {
				t.Errorf("Convert() mismatch: (-want +got):\n%s", diff)
			}
		})
	}
}
