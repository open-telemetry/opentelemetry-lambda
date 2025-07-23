package awstelemetryapi

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer/consumertest"
	"go.opentelemetry.io/collector/receiver/receivertest"
)

func TestFactory(t *testing.T) {
	factory := NewFactory("test-id")
	require.Equal(t, component.MustNewType(typeStr), factory.Type())

	// Test default config creation
	expectedCfg := &Config{
		extensionID: "test-id",
		Port:        defaultPort,
		Types:       []string{platform, function, extension},
		MaxItems:    defaultMaxItems,
		MaxBytes:    defaultMaxBytes,
		TimeoutMS:   defaultTimeoutMS,
	}
	require.Equal(t, expectedCfg, factory.CreateDefaultConfig())

	nopSettings := receivertest.NewNopSettings(component.MustNewType(typeStr))
	nopConsumer := consumertest.NewNop()

	// Test logs receiver creation
	_, err := factory.CreateLogs(
		context.Background(),
		nopSettings,
		factory.CreateDefaultConfig(),
		nopConsumer,
	)
	require.NoError(t, err)

	// Test traces receiver creation
	_, err = factory.CreateTraces(
		context.Background(),
		nopSettings,
		factory.CreateDefaultConfig(),
		nopConsumer,
	)
	require.NoError(t, err)

	// Test metrics receiver creation
	_, err = factory.CreateMetrics(
		context.Background(),
		nopSettings,
		factory.CreateDefaultConfig(),
		nopConsumer,
	)
	require.NoError(t, err)
}
