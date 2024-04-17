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
		// This test is first, because it illustrates the difference in making the rule that when
		// batch is present the converter appends decouple processor to the end of chain versus
		// the approach of this code which is to do this only when the last instance of batch
		// is not followed by decouple processor.
		{
			name: 	   "batch then decouple in middle of chain",
			processors: []interface{}{"processor1", "batch", "decouple", "processor2"},
			expectedProcessors: []interface{}{"processor1", "batch", "decouple", "processor2"},
		},
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
            name:       "no processors in chain",
            processors: nil,
            expectedProcessors: nil,
        },
        {
            name:       "batch processor in singleton chain",
            processors: []interface{}{"batch"},
            expectedProcessors: []interface{}{"batch", "decouple"},
        },
        {
            name:       "batch processor present twice",
            processors: []interface{}{"batch", "processor1", "batch"},
            expectedProcessors: []interface{}{"batch", "processor1", "batch", "decouple"},
        },

        {
            name:       "batch processor not present",
            processors: []interface{}{"processor1", "processor2"},
            expectedProcessors: []interface{}{"processor1", "processor2"},
        },
        {
            name:       "batch sandwiched between processors no decouple",
            processors: []interface{}{"processor1", "batch", "processor2"},
            expectedProcessors: []interface{}{"processor1", "batch", "processor2", "decouple"},
        },

        {
            name:       "batch and decouple processors already present in correct position",
            processors: []interface{}{"processor1", "batch", "processor2", "decouple"},
            expectedProcessors: []interface{}{"processor1", "batch", "processor2", "decouple"},
        },
        {
            name:       "decouple and batch",
            processors: []interface{}{"decouple", "batch"},
            expectedProcessors: []interface{}{"decouple", "batch", "decouple"},
        },
        {
            name:       "decouple then batch mixed with others in the pipelinefirst then batch somewhere",
            processors: []interface{}{"processor1", "decouple", "processor2", "batch", "processor3"},
            expectedProcessors: []interface{}{"processor1", "decouple", "processor2", "batch", "processor3", "decouple"},
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
