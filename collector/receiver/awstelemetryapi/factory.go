package awstelemetryapi

import (
	"context"
	"errors"

	"github.com/open-telemetry/opentelemetry-lambda/collector/receiver/awstelemetryapi/internal/sharedcomponent"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/receiver"
)

const (
	typeStr     = "awstelemetryapi"
	stability   = component.StabilityLevelBeta
	defaultPort = 4325
	platform    = "platform"
	function    = "function"
	extension   = "extension"
)

var receivers = sharedcomponent.NewSharedComponents()
var errConfigNotTelemetryAPI = errors.New("config was not an awstelemetryapi receiver config")

// NewFactory creates a new receiver factory.
func NewFactory(extensionID string) receiver.Factory {
	return receiver.NewFactory(
		component.MustNewType(typeStr),
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
	cfg, ok := rConf.(*Config)
	if !ok {
		return nil, errConfigNotTelemetryAPI
	}
	r := receivers.GetOrAdd(cfg, func() component.Component {
		// Use the new constructor name
		t, _ := newTelemetryAPIReceiver(cfg, params)
		return t
	})
	// Use the new struct name for the type assertion
	r.Unwrap().(*telemetryAPIReceiver).registerTracesConsumer(next)
	return r, nil
}

func createLogsReceiver(ctx context.Context, params receiver.Settings, rConf component.Config, next consumer.Logs) (receiver.Logs, error) {
	cfg, ok := rConf.(*Config)
	if !ok {
		return nil, errConfigNotTelemetryAPI
	}
	r := receivers.GetOrAdd(cfg, func() component.Component {
		t, _ := newTelemetryAPIReceiver(cfg, params)
		return t
	})
	r.Unwrap().(*telemetryAPIReceiver).registerLogsConsumer(next)
	return r, nil
}

func createMetricsReceiver(ctx context.Context, params receiver.Settings, rConf component.Config, next consumer.Metrics) (receiver.Metrics, error) {
	cfg, ok := rConf.(*Config)
	if !ok {
		return nil, errConfigNotTelemetryAPI
	}
	r := receivers.GetOrAdd(cfg, func() component.Component {
		t, _ := newTelemetryAPIReceiver(cfg, params)
		return t
	})
	r.Unwrap().(*telemetryAPIReceiver).registerMetricsConsumer(next)
	return r, nil
}
