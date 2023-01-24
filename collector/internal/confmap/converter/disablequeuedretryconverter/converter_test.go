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

package disablequeuedretryconverter

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/collector/confmap"
)

func TestConvert(t *testing.T) {
	for _, tc := range []struct {
		name     string
		conf     *confmap.Conf
		expected *confmap.Conf
		err      error
	}{
		{
			name:     "no exporters",
			conf:     confmap.New(),
			expected: confmap.New(),
			err:      nil,
		},
		{
			name:     "no queuing exporters",
			conf:     confmap.NewFromStringMap(map[string]any{"exporters": map[string]any{"prometheus": map[string]any{}}}),
			expected: confmap.NewFromStringMap(map[string]any{"exporters": map[string]any{"prometheus": map[string]any{}}}),
			err:      nil,
		},
		{
			name:     "some queuing exporters",
			conf:     confmap.NewFromStringMap(map[string]any{"exporters": map[string]any{"prometheus": map[string]any{}, "otlp": map[string]any{}}}),
			expected: confmap.NewFromStringMap(map[string]any{"exporters": map[string]any{"prometheus": map[string]any{}, "otlp": map[string]any{"sending_queue": map[string]any{"enabled": false}}}}),
			err:      nil,
		},
		{
			name:     "many queuing exporters",
			conf:     confmap.NewFromStringMap(map[string]any{"exporters": map[string]any{"otlphttp": map[string]any{}, "otlp": map[string]any{}}}),
			expected: confmap.NewFromStringMap(map[string]any{"exporters": map[string]any{"otlphttp": map[string]any{"sending_queue": map[string]any{"enabled": false}}, "otlp": map[string]any{"sending_queue": map[string]any{"enabled": false}}}}),
			err:      nil,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := New()
			err := c.Convert(context.Background(), tc.conf)
			assert.Equal(t, err, tc.err)
			assert.Equal(t, tc.conf, tc.expected)
		})
	}
}
