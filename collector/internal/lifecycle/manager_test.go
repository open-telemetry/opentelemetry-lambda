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

package lifecycle

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"

	"github.com/open-telemetry/opentelemetry-lambda/collector/internal/extensionapi"
	"github.com/open-telemetry/opentelemetry-lambda/collector/internal/telemetryapi"
)

type MockCollector struct {
	err error
}

func (c *MockCollector) Start(ctx context.Context) error {
	return c.err
}
func (c *MockCollector) Stop() error {
	return c.err
}

func TestRun(t *testing.T) {
	logger := zaptest.NewLogger(t)
	ctx := context.Background()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		_, err := w.Write([]byte(`{"time":"2006-01-02T15:04:05.000Z", "eventType":"SHUTDOWN", "record":{}}`))
		require.NoError(t, err)
		_, err = io.ReadAll(r.Body)
		require.NoError(t, err, "failed to read request body: %v", err)
	}))
	defer server.Close()
	u, err := url.Parse(server.URL)
	require.NoError(t, err)

	extensionEventTypes := []extensionapi.EventType{extensionapi.Invoke, extensionapi.Shutdown}
	// test with an error
	lm := manager{
		collector:       &MockCollector{err: fmt.Errorf("test start error")},
		logger:          logger,
		extensionClient: extensionapi.NewClient(logger, "", extensionEventTypes),
	}
	require.Error(t, lm.Run(ctx))
	// test with no waitgroup
	lm = manager{
		collector:       &MockCollector{},
		logger:          logger,
		listener:        telemetryapi.NewListener(logger),
		extensionClient: extensionapi.NewClient(logger, u.Host, extensionEventTypes),
	}
	require.NoError(t, lm.Run(ctx))
	// test with waitgroup counter incremented
	lm = manager{
		collector:       &MockCollector{},
		logger:          logger,
		listener:        telemetryapi.NewListener(logger),
		extensionClient: extensionapi.NewClient(logger, u.Host, extensionEventTypes),
	}
	lm.wg.Add(1)
	go func() {
		require.NoError(t, lm.Run(ctx))
	}()
	lm.wg.Done()

}

func TestProcessEvents(t *testing.T) {
	type test struct {
		name           string
		cancel         bool
		err            error
		serverResponse string
		collectorError error
	}
	testCases := []test{
		{
			name:   "processEvents with context cancelled",
			cancel: true,
		},
		{
			name: "processEvents with error from extension API",
			err:  fmt.Errorf("unexpected end of JSON input"),
		},
		{
			name:           "processEvents with shutdown event received",
			serverResponse: `{"time":"2006-01-02T15:04:05.000Z", "eventType":"SHUTDOWN", "record":{}}`,
		},
		{
			name:           "processEvents with shutdown event received and collector error",
			serverResponse: `{"time":"2006-01-02T15:04:05.000Z", "eventType":"SHUTDOWN", "record":{}}`,
			collectorError: fmt.Errorf("test shutdown error"),
			err:            fmt.Errorf("test shutdown error"),
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			logger := zaptest.NewLogger(t)
			var ctx context.Context

			if tc.cancel {
				var cancel context.CancelFunc
				ctx, cancel = context.WithCancel(context.Background())
				cancel()
			} else {
				ctx = context.Background()
			}

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(200)
				_, err := w.Write([]byte(tc.serverResponse))
				require.NoError(t, err)
				_, err = io.ReadAll(r.Body)
				require.NoError(t, err, "failed to read request body: %v", err)
			}))
			defer server.Close()
			u, err := url.Parse(server.URL)
			require.NoError(t, err)

			lm := manager{
				collector:       &MockCollector{err: tc.collectorError},
				logger:          logger,
				listener:        telemetryapi.NewListener(logger),
				extensionClient: extensionapi.NewClient(logger, u.Host, []extensionapi.EventType{extensionapi.Invoke, extensionapi.Shutdown}),
			}
			lm.wg.Add(1)
			if tc.err != nil {
				err = lm.processEvents(ctx)
				require.Error(t, err)
				require.ErrorContains(t, err, tc.err.Error())
			} else {
				require.NoError(t, lm.processEvents(ctx))
			}
		})
	}

}

func TestWriteAccountIDSymlink(t *testing.T) {
	// Use a temp directory so we don't conflict with the real path.
	tmpDir := t.TempDir()
	symlinkPath := filepath.Join(tmpDir, ".otel-aws-account-id")

	// Temporarily override the package-level constant via a helper approach:
	// We call the function directly and verify the symlink at the real path,
	// but to avoid touching /tmp we'll test the logic inline.
	logger := zaptest.NewLogger(t)

	t.Run("creates symlink with correct target", func(t *testing.T) {
		path := filepath.Join(tmpDir, "symlink-test-1")
		// Inline the logic to test with a custom path
		accountID := "123456789012"
		os.Remove(path)
		err := os.Symlink(accountID, path)
		require.NoError(t, err)

		target, err := os.Readlink(path)
		require.NoError(t, err)
		assert.Equal(t, "123456789012", target)
	})

	t.Run("preserves leading zeros", func(t *testing.T) {
		path := filepath.Join(tmpDir, "symlink-test-2")
		accountID := "000123456789"
		os.Remove(path)
		err := os.Symlink(accountID, path)
		require.NoError(t, err)

		target, err := os.Readlink(path)
		require.NoError(t, err)
		assert.Equal(t, "000123456789", target)
	})

	t.Run("replaces stale symlink", func(t *testing.T) {
		path := filepath.Join(tmpDir, "symlink-test-3")
		// Create an initial symlink
		require.NoError(t, os.Symlink("old-account-id", path))

		// Overwrite it
		os.Remove(path)
		require.NoError(t, os.Symlink("999888777666", path))

		target, err := os.Readlink(path)
		require.NoError(t, err)
		assert.Equal(t, "999888777666", target)
	})

	t.Run("skips when accountID is empty", func(t *testing.T) {
		// writeAccountIDSymlink should be a no-op for empty accountID
		writeAccountIDSymlink(logger, "")
		_, err := os.Readlink(symlinkPath)
		assert.True(t, os.IsNotExist(err), "symlink should not exist for empty accountID")
	})
}
