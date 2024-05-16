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

package telemetryapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"go.uber.org/zap"
)

const (
	ApiVersion20220701             = "2022-07-01"
	ApiVersionLatest               = ApiVersion20220701
	SchemaVersion20220701          = "2022-07-01"
	SchemaVersion20221213          = "2022-12-13"
	SchemaVersionLatest            = SchemaVersion20221213
	lambdaAgentIdentifierHeaderKey = "Lambda-Extension-Identifier"
)

type Client struct {
	logger     *zap.Logger
	httpClient *http.Client
	baseURL    string
}

func NewClient(logger *zap.Logger) *Client {
	return &Client{
		logger:     logger.Named("telemetryAPI.Client"),
		httpClient: &http.Client{},
		baseURL:    fmt.Sprintf("http://%s/%s/telemetry", os.Getenv("AWS_LAMBDA_RUNTIME_API"), ApiVersionLatest),
	}
}

func (c *Client) SubscribeEvents(ctx context.Context, eventTypes []EventType, extensionID string, listenerURI string) (string, error) {
	bufferingConfig := BufferingCfg{
		MaxItems:  1000,
		MaxBytes:  256 * 1024,
		TimeoutMS: 25,
	}

	destination := Destination{
		Protocol:   HttpProto,
		HttpMethod: HttpPost,
		Encoding:   JSON,
		URI:        URI(listenerURI),
	}

	data, err := json.Marshal(
		&SubscribeRequest{
			SchemaVersion: SchemaVersionLatest,
			EventTypes:    eventTypes,
			BufferingCfg:  bufferingConfig,
			Destination:   destination,
		})

	if err != nil {
		return "", fmt.Errorf("Failed to marshal SubscribeRequest: %w", err)
	}

	headers := make(map[string]string)
	headers[lambdaAgentIdentifierHeaderKey] = extensionID

	c.logger.Info("Subscribing", zap.String("baseURL", c.baseURL))
	resp, err := httpPutWithHeaders(ctx, c.httpClient, c.baseURL, data, headers)
	if err != nil {
		c.logger.Error("Subscription failed", zap.Error(err))
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusAccepted {
		c.logger.Error("Subscription failed. Logs API is not supported! Is this extension running in a local sandbox?", zap.Int("status_code", resp.StatusCode))
	} else if resp.StatusCode != http.StatusOK {
		c.logger.Error("Subscription failed")
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return "", fmt.Errorf("request to %s failed: %d[%s]: %w", c.baseURL, resp.StatusCode, resp.Status, err)
		}

		return "", fmt.Errorf("request to %s failed: %d[%s] %s", c.baseURL, resp.StatusCode, resp.Status, string(body))
	}

	body, _ := io.ReadAll(resp.Body)
	c.logger.Info("Subscription success", zap.String("response", string(body)))

	return string(body), nil
}

func (c *Client) Subscribe(ctx context.Context, extensionID string, listenerURI string) (string, error) {
	eventTypes := []EventType{
		Platform,
		// Function,
		// Extension,
	}
	return c.SubscribeEvents(ctx, eventTypes, extensionID, listenerURI)
}

func httpPutWithHeaders(ctx context.Context, client *http.Client, url string, data []byte, headers map[string]string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, "PUT", url, bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}

	contentType := "application/json"
	req.Header.Set("Content-Type", contentType)
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	return client.Do(req)
}
