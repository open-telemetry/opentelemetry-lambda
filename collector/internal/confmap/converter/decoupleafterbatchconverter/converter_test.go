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
package decoupleafterbatchconverter

import (
	"context"
	"testing"

	"go.opentelemetry.io/collector/confmap"

	"github.com/google/go-cmp/cmp"
)

func TestConvert(t *testing.T) {
	// Since this really tests differences in processors, it's easier to read cases
	// without the repeated definition of other fields in the config.
    baseConf := func(processors []interface{}) *confmap.Conf {
        return confmap.NewFromStringMap(map[string]interface{}{
            "service": map[string]interface{}{
                "pipelines": map[string]interface{}{
                    "traces": map[string]interface{}{
                        "processors": processors,
                    },
                },
            },
        })
    }

    testCases := []struct {
        name        string
        processors  []interface{}
        expectedProcessors []interface{}
        err         error
    }{
        {
            name:       "no service",
            processors: nil,
            expectedProcessors: nil,
        },
        {
            name:       "no pipelines",
            processors: nil,
            expectedProcessors: nil,
        },
        {
            name:       "no processors in pipeline",
            processors: nil,
            expectedProcessors: nil,
        },
        {
            name:       "batch processor present",
            processors: []interface{}{"batch"},
            expectedProcessors: []interface{}{"batch", "decouple"},
        },
        {
            name:       "batch processor not present",
            processors: []interface{}{"processor1", "processor2"},
            expectedProcessors: []interface{}{"processor1", "processor2"},
        },
        {
            name:       "batch and decouple processors already present",
            processors: []interface{}{"processor1", "batch", "processor2", "decouple"},
            expectedProcessors: []interface{}{"processor1", "batch", "processor2", "decouple"},
        },
    }

    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            conf := baseConf(tc.processors)
            expected := baseConf(tc.expectedProcessors)

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
