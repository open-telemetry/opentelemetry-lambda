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

package extensionapi

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegisterResponseUnmarshalJSON(t *testing.T) {
	tests := []struct {
		name        string
		jsonData    string
		expectedID  string
		shouldError bool
	}{
		{
			name: "standard_account_id",
			jsonData: `{
				"functionName": "test-function",
				"functionVersion": "$LATEST",
				"handler": "index.handler",
				"accountId": "123456789012"
			}`,
			expectedID:  "123456789012",
			shouldError: false,
		},
		{
			name: "account_id_with_leading_zero",
			jsonData: `{
				"functionName": "test-function",
				"functionVersion": "$LATEST",
				"handler": "index.handler",
				"accountId": "012345678901"
			}`,
			expectedID:  "012345678901",
			shouldError: false,
		},
		{
			name: "account_id_with_multiple_leading_zeros",
			jsonData: `{
				"functionName": "test-function",
				"functionVersion": "$LATEST",
				"handler": "index.handler",
				"accountId": "001234567890"
			}`,
			expectedID:  "001234567890",
			shouldError: false,
		},
		{
			name: "all_zeros_account_id",
			jsonData: `{
				"functionName": "test-function",
				"functionVersion": "$LATEST",
				"handler": "index.handler",
				"accountId": "000000000000"
			}`,
			expectedID:  "000000000000",
			shouldError: false,
		},
		{
			name: "missing_account_id",
			jsonData: `{
				"functionName": "test-function",
				"functionVersion": "$LATEST",
				"handler": "index.handler"
			}`,
			expectedID:  "",
			shouldError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var resp RegisterResponse
			err := json.Unmarshal([]byte(tt.jsonData), &resp)

			if tt.shouldError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedID, resp.AccountId, "AccountId should match exactly (leading zeros preserved)")
				assert.Equal(t, "test-function", resp.FunctionName)
				assert.Equal(t, "$LATEST", resp.FunctionVersion)
				assert.Equal(t, "index.handler", resp.Handler)
			}
		})
	}
}

func TestRegisterResponseLeadingZerosPreserved(t *testing.T) {
	// This test specifically validates that leading zeros are preserved
	// through the entire JSON unmarshaling process
	jsonData := `{
		"functionName": "my-function",
		"functionVersion": "1",
		"handler": "handler.main",
		"accountId": "012345678901"
	}`

	var resp RegisterResponse
	err := json.Unmarshal([]byte(jsonData), &resp)
	require.NoError(t, err)

	// Verify leading zero is preserved
	assert.Equal(t, "012345678901", resp.AccountId)
	assert.Len(t, resp.AccountId, 12, "Account ID should be exactly 12 characters")

	// Verify it's a string, not converted to a number
	assert.IsType(t, "", resp.AccountId)
}
