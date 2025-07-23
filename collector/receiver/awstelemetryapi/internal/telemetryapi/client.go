package telemetryapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"go.uber.org/zap"
)

// Client is a client for the AWS Lambda Telemetry API.
type Client struct {
	httpClient *http.Client
	baseURL    string
	logger     *zap.Logger
}

// NewClient creates a new Telemetry API client.
func NewClient(logger *zap.Logger) (*Client, error) {
	baseURL, ok := os.LookupEnv("AWS_LAMBDA_RUNTIME_API")
	if !ok {
		return nil, fmt.Errorf("AWS_LAMBDA_RUNTIME_API environment variable not set")
	}

	return &Client{
		httpClient: &http.Client{},
		baseURL:    fmt.Sprintf("http://%s/2020-01-01", baseURL),
		logger:     logger,
	}, nil
}

// Register registers the extension with the Lambda Extensions API.
func (c *Client) Register(ctx context.Context) (string, error) {
	url := c.baseURL + "/extension/register"
	reqBody, _ := json.Marshal(RegisterRequest{Events: []string{"INVOKE", "SHUTDOWN"}})

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return "", err
	}
	req.Header.Set("Lambda-Extension-Name", "logzio-otel-awstelemetryapi-receiver")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to register extension, status: %s", resp.Status)
	}

	extensionID := resp.Header.Get("Lambda-Extension-Identifier")
	if extensionID == "" {
		return "", fmt.Errorf("did not receive extension identifier")
	}

	return extensionID, nil
}

// Subscribe subscribes the extension to the Telemetry API.
func (c *Client) Subscribe(ctx context.Context, extensionID string, types []EventType, buffering BufferingCfg, destination Destination) error {
	url := fmt.Sprintf("http://%s/2022-07-01/telemetry", os.Getenv("AWS_LAMBDA_RUNTIME_API"))

	reqBody, _ := json.Marshal(SubscribeRequest{
		SchemaVersion: "2022-12-13",
		Types:         types,
		Buffering:     buffering,
		Destination:   destination,
	})

	req, err := http.NewRequestWithContext(ctx, "PUT", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return err
	}
	req.Header.Set("Lambda-Extension-Identifier", extensionID)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to subscribe to telemetry api, status: %s", resp.Status)
	}

	c.logger.Info("Successfully subscribed to Telemetry API", zap.Any("types", types))
	return nil
}
