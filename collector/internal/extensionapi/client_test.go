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

package extensionapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

func TestRegisterSendsAcceptFeatureHeader(t *testing.T) {
	var receivedAcceptFeature string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedAcceptFeature = r.Header.Get("Lambda-Extension-Accept-Feature")
		w.Header().Set("Lambda-Extension-Identifier", "test-ext-id")
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"functionName":"my-func","functionVersion":"$LATEST","handler":"index.handler","accountId":"123456789012"}`))
	}))
	defer server.Close()

	u, err := url.Parse(server.URL)
	require.NoError(t, err)

	logger := zaptest.NewLogger(t)
	// The client prepends "http://" and appends "/2020-01-01/extension", so we
	// need to set up the server path accordingly. Instead, construct the client
	// with an empty base and override.
	client := NewClient(logger, u.Host)
	resp, err := client.Register(context.Background(), "test-extension")
	require.NoError(t, err)

	assert.Equal(t, "accountId", receivedAcceptFeature)
	assert.Equal(t, "123456789012", resp.AccountID)
	assert.Equal(t, "my-func", resp.FunctionName)
	assert.Equal(t, "test-ext-id", resp.ExtensionID)
}

func TestRegisterParsesAccountIDWithLeadingZeros(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Lambda-Extension-Identifier", "ext-id")
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"functionName":"f","functionVersion":"v","handler":"h","accountId":"000123456789"}`))
	}))
	defer server.Close()

	u, err := url.Parse(server.URL)
	require.NoError(t, err)

	logger := zaptest.NewLogger(t)
	client := NewClient(logger, u.Host)
	resp, err := client.Register(context.Background(), "test-extension")
	require.NoError(t, err)

	assert.Equal(t, "000123456789", resp.AccountID, "leading zeros must be preserved")
}
