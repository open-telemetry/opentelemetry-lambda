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

package deprecatedprocessorconverter

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/confmap"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

func TestConvert(t *testing.T) {
	for _, tc := range []struct {
		name     string
		conf     map[string]any
		expected []string
	}{
		{
			name:     "no processors",
			conf:     map[string]any{},
			expected: nil,
		},
		{
			name:     "no deprecated processors",
			conf:     map[string]any{"processors": map[string]any{"transform": map[string]any{}, "filter": nil}},
			expected: nil,
		},
		{
			name: "deprecated processors with and without instance names",
			conf: map[string]any{"processors": map[string]any{
				"transform":       map[string]any{},
				"span":            nil,
				"resource/env":    map[string]any{},
				"attributes/drop": map[string]any{},
				"attributes":      map[string]any{},
			}},
			expected: []string{"attributes", "attributes/drop", "resource/env", "span"},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			core, logs := observer.New(zapcore.WarnLevel)
			conf := confmap.NewFromStringMap(tc.conf)
			before := conf.ToStringMap()

			require.NoError(t, New(zap.New(core)).Convert(context.Background(), conf))

			assert.Equal(t, before, conf.ToStringMap(), "converter must not modify the config")

			var warned []string
			for _, entry := range logs.All() {
				assert.Equal(t, zapcore.WarnLevel, entry.Level)
				warned = append(warned, entry.ContextMap()["processor"].(string))
			}
			assert.Equal(t, tc.expected, warned)
		})
	}
}
