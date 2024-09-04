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

package telemetryapireceiver // import "github.com/open-telemetry/opentelemetry-lambda/collector/receiver/telemetryapireceiver"

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/confmap/confmaptest"
)

func TestLoadConfig(t *testing.T) {
	t.Parallel()

	// Helper function to create expected Config
	createExpectedConfig := func(types []string) *Config {
		return &Config{
			extensionID: "extensionID",
			Port:        12345,
			Types:       types,
		}
	}

	tests := []struct {
		name     string
		id       component.ID
		expected component.Config
	}{
		{
			name:     "default",
			id:       component.NewID(component.MustNewType("telemetryapi")),
			expected: NewFactory("extensionID").CreateDefaultConfig(),
		},
		{
			name:     "all types",
			id:       component.NewIDWithName(component.MustNewType("telemetryapi"), "1"),
			expected: createExpectedConfig([]string{platform, function, extension}),
		},
		{
			name:     "platform only",
			id:       component.NewIDWithName(component.MustNewType("telemetryapi"), "2"),
			expected: createExpectedConfig([]string{platform}),
		},
		{
			name:     "function only",
			id:       component.NewIDWithName(component.MustNewType("telemetryapi"), "3"),
			expected: createExpectedConfig([]string{function}),
		},
		{
			name:     "extension only",
			id:       component.NewIDWithName(component.MustNewType("telemetryapi"), "4"),
			expected: createExpectedConfig([]string{extension}),
		},
		{
			name:     "platform and function",
			id:       component.NewIDWithName(component.MustNewType("telemetryapi"), "5"),
			expected: createExpectedConfig([]string{platform, function}),
		},
		{
			name:     "platform and extension",
			id:       component.NewIDWithName(component.MustNewType("telemetryapi"), "6"),
			expected: createExpectedConfig([]string{platform, extension}),
		},
		{
			name:     "function and extension",
			id:       component.NewIDWithName(component.MustNewType("telemetryapi"), "7"),
			expected: createExpectedConfig([]string{function, extension}),
		},
		{
			name:     "empty types",
			id:       component.NewIDWithName(component.MustNewType("telemetryapi"), "8"),
			expected: createExpectedConfig([]string{}),
		},
		{
			name:     "function and extension (alternative syntax)",
			id:       component.NewIDWithName(component.MustNewType("telemetryapi"), "9"),
			expected: createExpectedConfig([]string{function, extension}),
		},
		{
			name:     "function and extension (another syntax)",
			id:       component.NewIDWithName(component.MustNewType("telemetryapi"), "10"),
			expected: createExpectedConfig([]string{function, extension}),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cm, err := confmaptest.LoadConf(filepath.Join("testdata", "config.yaml"))
			require.NoError(t, err)

			factory := NewFactory("extensionID")
			cfg := factory.CreateDefaultConfig()

			sub, err := cm.Sub(tt.id.String())
			require.NoError(t, err)
			require.NoError(t, sub.Unmarshal(cfg))
			require.NoError(t, component.ValidateConfig(cfg))

			require.Equal(t, tt.expected, cfg)
		})
	}
}

func TestValidate(t *testing.T) {
	testCases := []struct {
		desc        string
		cfg         *Config
		expectedErr error
	}{
		{
			desc:        "valid config",
			cfg:         &Config{},
			expectedErr: nil,
		},
		{
			desc: "invalid config",
			cfg: &Config{
				Types: []string{"invalid"},
			},
			expectedErr: fmt.Errorf("unknown extension type: invalid"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			actualErr := tc.cfg.Validate()
			if tc.expectedErr != nil {
				require.EqualError(t, actualErr, tc.expectedErr.Error())
			} else {
				require.NoError(t, actualErr)
			}

		})
	}
}
