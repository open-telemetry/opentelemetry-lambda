package solarwindsapmsettingsextension

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/extension"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

func TestCreateExtension(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		cfg  *Config
	}{
		{
			name: "default",
			cfg: &Config{
				Interval: time.Duration(10000000000),
			},
		},
		{
			name: "anything",
			cfg: &Config{
				Endpoint: "0.0.0.0:1234",
				Key:      "something",
				Interval: time.Duration(10000000000),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ex := createAnExtension(tt.cfg, t)
			require.NoError(t, ex.Shutdown(context.TODO()))
		})
	}
}

// create extension
func createAnExtension(c *Config, t *testing.T) extension.Extension {
	logger, err := zap.NewProduction()
	ex, err := newSolarwindsApmSettingsExtension(c, logger)
	require.NoError(t, err)
	err = ex.Start(context.TODO(), nil)
	require.NoError(t, err)
	return ex
}

func TestValidateSolarwindsApmSettingsExtensionConfiguration(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		cfg     *Config
		ok      bool
		message []string
	}{
		{
			name:    "nothing",
			cfg:     &Config{},
			ok:      false,
			message: []string{"endpoint must not be empty"},
		},
		{
			name: "valid configuration",
			cfg: &Config{
				Endpoint: "apm.collector.na-02.cloud.solarwinds.com:443",
				Key:      "token:name",
				Interval: time.Duration(10000000000),
			},
			ok:      true,
			message: []string{},
		},
		{
			name: "endpoint without :",
			cfg: &Config{
				Endpoint: "apm.collector.na-01.cloud.solarwinds.com",
			},
			ok:      false,
			message: []string{"endpoint should be in \"<host>:<port>\" format"},
		},
		{
			name: "endpoint with some :",
			cfg: &Config{
				Endpoint: "apm.collector.na-01.cloud.solarwinds.com:a:b",
			},
			ok:      false,
			message: []string{"endpoint should be in \"<host>:<port>\" format"},
		},
		{
			name: "endpoint with invalid port",
			cfg: &Config{
				Endpoint: "apm.collector.na-01.cloud.solarwinds.com:port",
			},
			ok:      false,
			message: []string{"the <port> portion of endpoint has to be an integer"},
		},
		{
			name: "bad endpoint",
			cfg: &Config{
				Endpoint: "apm.collector..cloud.solarwinds.com:443",
			},
			ok:      false,
			message: []string{"endpoint \"<host>\" part should be in \"apm.collector.[a-z]{2,3}-[0-9]{2}.[a-z\\-]*.solarwinds.com\" regex format, see https://documentation.solarwinds.com/en/success_center/observability/content/system_requirements/endpoints.htm for detail"},
		},
		{
			name: "empty endpoint with port",
			cfg: &Config{
				Endpoint: ":433",
			},
			ok:      false,
			message: []string{"endpoint should be in \"<host>:<port>\" format and \"<host>\" must not be empty"},
		},
		{
			name: "empty endpoint without port",
			cfg: &Config{
				Endpoint: ":",
			},
			ok:      false,
			message: []string{"endpoint should be in \"<host>:<port>\" format and \"<host>\" must not be empty"},
		},
		{
			name: "endpoint without port",
			cfg: &Config{
				Endpoint: "apm.collector.na-01.cloud.solarwinds.com:",
			},
			ok:      false,
			message: []string{"endpoint should be in \"<host>:<port>\" format and \"<port>\" must not be empty"},
		},
		{
			name: "valid endpoint but empty key",
			cfg: &Config{
				Endpoint: "apm.collector.na-01.cloud.solarwinds.com:443",
			},
			ok:      false,
			message: []string{"key must not be empty"},
		},
		{
			name: "key is :",
			cfg: &Config{
				Endpoint: "apm.collector.na-01.cloud.solarwinds.com:443",
				Key:      ":",
			},
			ok:      false,
			message: []string{"key should be in \"<token>:<service_name>\" format and \"<token>\" must not be empty"},
		},
		{
			name: "key is ::",
			cfg: &Config{
				Endpoint: "apm.collector.na-01.cloud.solarwinds.com:443",
				Key:      "::",
			},
			ok:      false,
			message: []string{"key should be in \"<token>:<service_name>\" format"},
		},
		{
			name: "key is :name",
			cfg: &Config{
				Endpoint: "apm.collector.na-01.cloud.solarwinds.com:443",
				Key:      ":name",
			},
			ok:      false,
			message: []string{"key should be in \"<token>:<service_name>\" format and \"<token>\" must not be empty"},
		},
		{
			name: "key is token:",
			cfg: &Config{
				Endpoint: "apm.collector.na-01.cloud.solarwinds.com:443",
				Key:      "token:",
			},
			ok:      false,
			message: []string{"<service_name> from config is empty. Trying to resolve service name from env variables using best effort", "Unable to resolve service name by our best effort. It can be defined via environment variables \"OTEL_SERVICE_NAME\" or \"AWS_LAMBDA_FUNCTION_NAME\"", "key should be in \"<token>:<service_name>\" format and \"<service_name>\" must not be empty"},
		},
		{
			name: "minimum_interval",
			cfg: &Config{
				Endpoint: "apm.collector.na-01.cloud.solarwinds.com:443",
				Key:      "token:name",
				Interval: time.Duration(4000000000),
			},
			ok:      true,
			message: []string{"Interval 4s is less than the minimum supported interval " + MinimumInterval.String() + ". use minimum interval " + MinimumInterval.String() + " instead"},
		},
		{
			name: "maximum_interval",
			cfg: &Config{
				Endpoint: "apm.collector.na-01.cloud.solarwinds.com:443",
				Key:      "token:name",
				Interval: time.Duration(61000000000),
			},
			ok:      true,
			message: []string{"Interval 1m1s is greater than the maximum supported interval " + MaximumInterval.String() + ". use maximum interval " + MaximumInterval.String() + " instead"},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			observedZapCore, observedLogs := observer.New(zap.DebugLevel)
			observedLogger := zap.New(observedZapCore)
			require.Equal(t, tc.ok, validateSolarwindsApmSettingsExtensionConfiguration(tc.cfg, observedLogger))
			require.Equal(t, len(tc.message), observedLogs.Len())
			for i, observedLog := range observedLogs.All() {
				require.Equal(t, tc.message[i], observedLog.Message)
			}
		})
	}
}

func TestResolveServiceNameBestEffort(t *testing.T) {
	// Without any environment variables
	require.Empty(t, resolveServiceNameBestEffort(zap.NewExample()))
	// With OTEL_SERVICE_NAME only
	require.NoError(t, os.Setenv("OTEL_SERVICE_NAME", "otel_ser1"))
	require.Equal(t, "otel_ser1", resolveServiceNameBestEffort(zap.NewExample()))
	require.NoError(t, os.Unsetenv("OTEL_SERVICE_NAME"))
	// With AWS_LAMBDA_FUNCTION_NAME only
	require.NoError(t, os.Setenv("AWS_LAMBDA_FUNCTION_NAME", "lambda"))
	require.Equal(t, "lambda", resolveServiceNameBestEffort(zap.NewExample()))
	require.NoError(t, os.Unsetenv("AWS_LAMBDA_FUNCTION_NAME"))
	// With both
	require.NoError(t, os.Setenv("OTEL_SERVICE_NAME", "otel_ser1"))
	require.NoError(t, os.Setenv("AWS_LAMBDA_FUNCTION_NAME", "lambda"))
	require.Equal(t, "otel_ser1", resolveServiceNameBestEffort(zap.NewExample()))
	require.NoError(t, os.Unsetenv("AWS_LAMBDA_FUNCTION_NAME"))
	require.NoError(t, os.Unsetenv("OTEL_SERVICE_NAME"))
}
