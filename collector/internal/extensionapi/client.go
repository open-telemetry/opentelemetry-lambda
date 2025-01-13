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
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"go.uber.org/zap"
)

// RegisterResponse is the body of the response for /register
type RegisterResponse struct {
	FunctionName    string `json:"functionName"`
	FunctionVersion string `json:"functionVersion"`
	Handler         string `json:"handler"`
	ExtensionID     string
}

// NextEventResponse is the response for /event/next
type NextEventResponse struct {
	EventType          EventType `json:"eventType"`
	DeadlineMs         int64     `json:"deadlineMs"`
	RequestID          string    `json:"requestId"`
	InvokedFunctionArn string    `json:"invokedFunctionArn"`
	Tracing            Tracing   `json:"tracing"`
}

// Tracing is part of the response for /event/next
type Tracing struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

// StatusResponse is the body of the response for /init/error and /exit/error
type StatusResponse struct {
	Status string `json:"status"`
}

// EventType represents the type of events recieved from /event/next
type EventType string

const (
	// Invoke is a lambda invoke
	Invoke EventType = "INVOKE"

	// Shutdown is a shutdown event for the environment
	Shutdown EventType = "SHUTDOWN"
)

const (
	extensionNameHeader      = "Lambda-Extension-Name"
	extensionIdentiferHeader = "Lambda-Extension-Identifier"
	extensionErrorType       = "Lambda-Extension-Function-Error-Type"
)

// Client is a simple client for the Lambda Extensions API.
type Client struct {
	baseURL     string
	httpClient  *http.Client
	extensionID string
	logger      *zap.Logger
}

// NewClient returns a Lambda Extensions API client.
func NewClient(logger *zap.Logger, awsLambdaRuntimeAPI string) *Client {
	baseURL := fmt.Sprintf("http://%s/2020-01-01/extension", awsLambdaRuntimeAPI)
	return &Client{
		baseURL:    baseURL,
		httpClient: &http.Client{},
		logger:     logger.Named("extensionAPI.Client"),
	}
}

// Register will register the extension with the Extensions API.
func (e *Client) Register(ctx context.Context, filename string) (*RegisterResponse, error) {
	const action = "/register"
	url := e.baseURL + action

	reqBody, err := json.Marshal(map[string]interface{}{
		"events": []EventType{Invoke, Shutdown},
	})
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set(extensionNameHeader, filename)

	var registerResp RegisterResponse
	resp, err := e.doRequest(req, &registerResp)
	if err != nil {
		return nil, err
	}
	e.extensionID = resp.Header.Get(extensionIdentiferHeader)
	e.logger.Debug("Registered extension", zap.String("ID", e.extensionID))

	registerResp.ExtensionID = e.extensionID
	return &registerResp, nil
}

// NextEvent blocks while long polling for the next lambda invoke or shutdown.
func (e *Client) NextEvent(ctx context.Context) (*NextEventResponse, error) {
	const action = "/event/next"
	url := e.baseURL + action

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set(extensionIdentiferHeader, e.extensionID)

	var nextEventResp NextEventResponse
	if _, err := e.doRequest(req, &nextEventResp); err != nil {
		return nil, err
	}
	return &nextEventResp, nil
}

// InitError reports an initialization error to the platform.
// Call it when you registered but failed to initialize.
func (e *Client) InitError(ctx context.Context, errorType string) (*StatusResponse, error) {
	const action = "/init/error"
	url := e.baseURL + action

	req, err := http.NewRequestWithContext(ctx, "POST", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set(extensionIdentiferHeader, e.extensionID)
	req.Header.Set(extensionErrorType, errorType)

	var statusResp StatusResponse
	if _, err := e.doRequest(req, &statusResp); err != nil {
		return nil, err
	}
	return &statusResp, nil
}

// ExitError reports an error to the platform before exiting.
// Call it when you encounter an unexpected failure.
func (e *Client) ExitError(ctx context.Context, errorType string) (*StatusResponse, error) {
	const action = "/exit/error"
	url := e.baseURL + action

	req, err := http.NewRequestWithContext(ctx, "POST", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set(extensionIdentiferHeader, e.extensionID)
	req.Header.Set(extensionErrorType, errorType)

	var statusResp StatusResponse
	if _, err := e.doRequest(req, &statusResp); err != nil {
		return nil, err
	}
	return &statusResp, nil
}

func (e *Client) doRequest(req *http.Request, out interface{}) (*http.Response, error) {
	resp, err := e.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return nil, fmt.Errorf("request failed with status %s", resp.Status)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if err = json.Unmarshal(body, out); err != nil {
		return resp, err
	}
	return resp, nil
}
