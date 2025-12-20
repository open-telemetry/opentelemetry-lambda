// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package lambdalifecycle

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestInitType_String(t *testing.T) {
	tests := []struct {
		initType InitType
		expected string
	}{
		{OnDemand, "on-demand"},
		{ProvisionedConcurrency, "provisioned-concurrency"},
		{SnapStart, "snap-start"},
		{LambdaManagedInstances, "lambda-managed-instance"},
		{Unknown, "unknown"},
		{InitType(99), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if result := tt.initType.String(); result != tt.expected {
				t.Errorf("InitType.String() = %q, expected %q", result, tt.expected)
			}
		})
	}
}

func TestParseInitType(t *testing.T) {
	tests := []struct {
		input    string
		expected InitType
	}{
		{"on-demand", OnDemand},
		{"provisioned-concurrency", ProvisionedConcurrency},
		{"snap-start", SnapStart},
		{"lambda-managed-instances", LambdaManagedInstances},
		{"unknown", Unknown},
		{"", Unknown},
		{"invalid", Unknown},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if result := ParseInitType(tt.input); result != tt.expected {
				t.Errorf("ParseInitType(%q) = %v, expected %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestInitTypeFromEnv(t *testing.T) {
	const testEnvVar = "TEST_INIT_TYPE"

	tests := []struct {
		name     string
		envVal   string
		expected InitType
		setEnv   bool
	}{
		{"on-demand", "on-demand", OnDemand, true},
		{"provisioned-concurrency", "provisioned-concurrency", ProvisionedConcurrency, true},
		{"snap-start", "snap-start", SnapStart, true},
		{"lambda-managed-instances", "lambda-managed-instances", LambdaManagedInstances, true},
		{"unset env var", "", Unknown, false},
		{"empty env var", "", Unknown, true},
		{"invalid value", "foo", Unknown, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.NoError(t, os.Unsetenv(testEnvVar))
			if tt.setEnv {
				require.NoError(t, os.Setenv(testEnvVar, tt.envVal))
			}

			if result := InitTypeFromEnv(testEnvVar); result != tt.expected {
				t.Errorf("InitTypeFromEnv() = %v, expected %v", result, tt.expected)
			}
		})
	}
}
