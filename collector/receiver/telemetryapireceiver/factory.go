package telemetryapireceiver

import (
	"context"
	"errors"
	"fmt"

	"github.com/open-telemetry/opentelemetry-lambda/collector/receiver/telemetryapireceiver/internal/sharedcomponent"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/receiver"
)

const (
	typeStr          = "telemetryapireceiver"
	stability        = component.StabilityLevelDevelopment
	platform         = "platform"
	function         = "function"
	extension        = "extension"
	defaultPort      = 4325
	defaultMaxItems  = 1000
	defaultMaxBytes  = 262144
	defaultTimeoutMS = 1000
)

var (
	Type                     = component.MustNewType(typeStr)
	errConfigNotTelemetryAPI = errors.New("config was not a Telemetry API receiver config")
)

// NewFactory creates a new receiver factory.
func NewFactory(extensionID string) receiver.Factory {
	return receiver.NewFactory(
		Type,
		func() component.Config {
			return &Config{
				extensionID: extensionID,
				Port:        defaultPort,
				Types:       []string{platform, function, extension},
				MaxItems:    defaultMaxItems,
				MaxBytes:    defaultMaxBytes,
				TimeoutMS:   defaultTimeoutMS,
			}
		},
		receiver.WithTraces(createTracesReceiver, stability),
		receiver.WithLogs(createLogsReceiver, stability),
		receiver.WithMetrics(createMetricsReceiver, stability),
	)
}

func createTracesReceiver(ctx context.Context, params receiver.Settings, rConf component.Config, next consumer.Traces) (receiver.Traces, error) {
	// Use a single helper function for creating/getting and wiring up the receiver.
	shared, err := getOrCreateReceiver(params, rConf)
	if err != nil {
		return nil, err
	}
	shared.Unwrap().(*telemetryAPIReceiver).registerTracesConsumer(next)
	return shared, nil
}

func createLogsReceiver(ctx context.Context, params receiver.Settings, rConf component.Config, next consumer.Logs) (receiver.Logs, error) {
	shared, err := getOrCreateReceiver(params, rConf)
	if err != nil {
		return nil, err
	}
	shared.Unwrap().(*telemetryAPIReceiver).registerLogsConsumer(next)
	return shared, nil
}

func createMetricsReceiver(ctx context.Context, params receiver.Settings, rConf component.Config, next consumer.Metrics) (receiver.Metrics, error) {
	shared, err := getOrCreateReceiver(params, rConf)
	if err != nil {
		return nil, err
	}
	shared.Unwrap().(*telemetryAPIReceiver).registerMetricsConsumer(next)
	return shared, nil
}

// getOrCreateReceiver handles the logic of creating or retrieving the shared receiver instance.
func getOrCreateReceiver(params receiver.Settings, rConf component.Config) (*sharedcomponent.SharedComponent, error) {
	cfg, ok := rConf.(*Config)
	if !ok {
		return nil, errConfigNotTelemetryAPI
	}

	var createdReceiver component.Component
	shared := receivers.GetOrAdd(cfg, func() component.Component {
		var err error
		createdReceiver, err = newTelemetryAPIReceiver(cfg, params)
		if err != nil {
			createdReceiver = nil
		}
		return createdReceiver
	})

	if shared.Unwrap() == nil {
		return nil, fmt.Errorf("failed to create telemetry API receiver")
	}

	return shared, nil
}

var receivers = sharedcomponent.NewSharedComponents()
