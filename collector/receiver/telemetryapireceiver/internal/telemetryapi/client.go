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
	awsLambdaRuntimeAPIEnvVar       = "AWS_LAMBDA_RUNTIME_API"
	lambdaExtensionNameHeader       = "Lambda-Extension-Name"
	lambdaExtensionIdentifierHeader = "Lambda-Extension-Identifier"
)

// Client is a client for the AWS Lambda Telemetry API.
type Client struct {
	httpClient      *http.Client
	baseURL         string
	telemetryAPIURL string
	logger          *zap.Logger
}

// NewClient creates a new Telemetry API client.
func NewClient(logger *zap.Logger) (*Client, error) {
	runtimeAPI, ok := os.LookupEnv(awsLambdaRuntimeAPIEnvVar)
	if !ok {
		return nil, fmt.Errorf("%s environment variable not set", awsLambdaRuntimeAPIEnvVar)
	}

	return &Client{
		httpClient:      &http.Client{},
		baseURL:         fmt.Sprintf("http://%s/2020-01-01", runtimeAPI),
		telemetryAPIURL: fmt.Sprintf("http://%s/2022-07-01/telemetry", runtimeAPI),
		logger:          logger,
	}, nil
}

// Register registers the extension with the Lambda Extensions API.
func (c *Client) Register(ctx context.Context, extensionName string) (string, error) {
	url := c.baseURL + "/extension/register"
	reqBody, _ := json.Marshal(RegisterRequest{Events: []string{"INVOKE", "SHUTDOWN"}})

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return "", err
	}
	req.Header.Set(lambdaExtensionNameHeader, extensionName)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("failed to register extension, status: %s, body: %s", resp.Status, string(body))
	}

	extensionID := resp.Header.Get(lambdaExtensionNameHeader)
	if extensionID == "" {
		return "", fmt.Errorf("did not receive extension identifier")
	}

	return extensionID, nil
}

// Subscribe subscribes the extension to the Telemetry API.
func (c *Client) Subscribe(ctx context.Context, extensionID string, types []EventType, buffering BufferingCfg, destination Destination) error {
	url := c.telemetryAPIURL

	reqBody, err := json.Marshal(SubscribeRequest{
		SchemaVersion: "2022-12-13",
		Types:         types,
		Buffering:     buffering,
		Destination:   destination,
	})
	if err != nil {
		return fmt.Errorf("failed to marshal subscribe request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "PUT", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return err
	}
	req.Header.Set(lambdaExtensionNameHeader, extensionID)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to subscribe to telemetry api, status: %s, body: %s", resp.Status, string(body))
	}

	c.logger.Info("Successfully subscribed to Telemetry API", zap.Any("types", types))
	return nil
}
