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

package decouplefirstconverter

import (
	"context"
	"testing"

	"go.opentelemetry.io/collector/confmap"

	"github.com/google/go-cmp/cmp"
)

func TestConvert(t *testing.T) {
	for _, tc := range []struct {
		name     string
		conf     *confmap.Conf
		expected *confmap.Conf
		err      error
	}{
		{
			name:     "no pipelines",
			conf:     confmap.New(),
			expected: confmap.New(),
			err:      nil,
		},
		{
			name: "no processors in pipeline",
			conf: confmap.NewFromStringMap(map[string]interface{}{
				"service": map[string]interface{}{
					"pipelines": map[string]interface{}{
						"traces": map[string]interface{}{},
					},
				},
			}),
			expected: confmap.NewFromStringMap(map[string]interface{}{
				"service": map[string]interface{}{
					"pipelines": map[string]interface{}{
						"traces": map[string]interface{}{},
					},
				},
			}),
			err: nil,
		},
		{
			name: "decouple processor first",
			conf: confmap.NewFromStringMap(map[string]interface{}{
				"service": map[string]interface{}{
					"pipelines": map[string]interface{}{
						"traces": map[string]interface{}{
							"processors": []interface{}{"decouple", "processor1"},
						},
					},
				},
			}),
			expected: confmap.NewFromStringMap(map[string]interface{}{
				"service": map[string]interface{}{
					"pipelines": map[string]interface{}{
						"traces": map[string]interface{}{
							"processors": []interface{}{"decouple", "processor1"},
						},
					},
				},
			}),
			err: nil,
		},
		{
			name: "decouple processor not first",
			conf: confmap.NewFromStringMap(map[string]interface{}{
				"service": map[string]interface{}{
					"pipelines": map[string]interface{}{
						"traces": map[string]interface{}{
							"processors": []interface{}{"processor1", "decouple", "processor2"},
						},
					},
				},
			}),
			expected: confmap.NewFromStringMap(map[string]interface{}{
				"service": map[string]interface{}{
					"pipelines": map[string]interface{}{
						"traces": map[string]interface{}{
							"processors": []interface{}{"decouple", "processor1", "processor2"},
						},
					},
				},
			}),
			err: nil,
		},
		{
			name: "multiple decouple processors",
			conf: confmap.NewFromStringMap(map[string]interface{}{
				"service": map[string]interface{}{
					"pipelines": map[string]interface{}{
						"traces": map[string]interface{}{
							"processors": []interface{}{"processor1", "decouple", "decouple/instance1", "processor2", "decouple/instance2"},
						},
					},
				},
			}),
			expected: confmap.NewFromStringMap(map[string]interface{}{
				"service": map[string]interface{}{
					"pipelines": map[string]interface{}{
						"traces": map[string]interface{}{
							"processors": []interface{}{"decouple", "processor1", "processor2"},
						},
					},
				},
			}),
			err: nil,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := New()
			if err := c.Convert(context.Background(), tc.conf); err != nil {
				t.Errorf("unexpected error converting: %v", err)
			}

			// check that tc.conf is equal to tc.expected, but for something
			// that must work on maps which are unordered
			// assert.NoError(t, err)
			if diff := cmp.Diff(tc.expected.ToStringMap(), tc.conf.ToStringMap()); diff != "" {
				t.Errorf("Convert() mismatch: (-want +got):\n%s", diff)
			}
		})
	}
}
