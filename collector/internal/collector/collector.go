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

package collector

import (
	"context"
	"fmt"
	"os"

	"github.com/open-telemetry/opentelemetry-collector-contrib/confmap/provider/s3provider"
	"github.com/open-telemetry/opentelemetry-collector-contrib/confmap/provider/secretsmanagerprovider"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/confmap"
	"go.opentelemetry.io/collector/confmap/converter/expandconverter"
	"go.opentelemetry.io/collector/confmap/provider/envprovider"
	"go.opentelemetry.io/collector/confmap/provider/fileprovider"
	"go.opentelemetry.io/collector/confmap/provider/httpprovider"
	"go.opentelemetry.io/collector/confmap/provider/yamlprovider"
	"go.opentelemetry.io/collector/otelcol"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/open-telemetry/opentelemetry-lambda/collector/internal/confmap/converter/disablequeuedretryconverter"
)

// Collector runs a single otelcol as a go routine within the
// same process as the executor.
type Collector struct {
	factories otelcol.Factories
	cfgProSet otelcol.ConfigProviderSettings
	svc       *otelcol.Collector
	appDone   chan struct{}
	stopped   bool
	logger    *zap.Logger
	version   string
}

func getConfig(logger *zap.Logger) string {
	val, ex := os.LookupEnv("OPENTELEMETRY_COLLECTOR_CONFIG_FILE")
	if !ex {
		return "/opt/collector-config/config.yaml"
	}
	logger.Info("Using config URI from environment", zap.String("uri", val))
	return val
}

func NewCollector(logger *zap.Logger, factories otelcol.Factories, version string) *Collector {
	l := logger.Named("NewCollector")
	cfgSet := otelcol.ConfigProviderSettings{
		ResolverSettings: confmap.ResolverSettings{
			URIs:              []string{getConfig(l)},
			ProviderFactories: []confmap.ProviderFactory{fileprovider.NewFactory(), envprovider.NewFactory(), yamlprovider.NewFactory(), httpprovider.NewFactory(), s3provider.NewFactory(), secretsmanagerprovider.NewFactory()},
			ConverterFactories: []confmap.ConverterFactory{
				expandconverter.NewFactory(),
				confmap.NewConverterFactory(func(set confmap.ConverterSettings) confmap.Converter {
					return disablequeuedretryconverter.New()
				}),
			},
		},
	}

	col := &Collector{
		factories: factories,
		cfgProSet: cfgSet,
		logger:    logger,
		version:   version,
	}
	return col
}

func (c *Collector) Start(ctx context.Context) error {
	params := otelcol.CollectorSettings{
		BuildInfo: component.BuildInfo{
			Command:     "otelcol-lambda",
			Description: "SolarWinds Observability (SWO) distribution of Lambda Collector",
			Version:     c.version,
		},
		ConfigProviderSettings: c.cfgProSet,
		Factories: func() (otelcol.Factories, error) {
			return c.factories, nil
		},
		LoggingOptions: []zap.Option{zap.WrapCore(func(_ zapcore.Core) zapcore.Core {
			return c.logger.Core()
		})},
	}
	var err error
	c.svc, err = otelcol.NewCollector(params)
	if err != nil {
		return err
	}

	c.appDone = make(chan struct{})

	go func() {
		defer close(c.appDone)
		appErr := c.svc.Run(ctx)
		if appErr != nil {
			err = appErr
		}
	}()

	for {
		state := c.svc.GetState()

		// While waiting for collector start, an error was found. Most likely
		// an invalid custom collector configuration file.
		if err != nil {
			return err
		}

		switch state {
		case otelcol.StateStarting:
			// NoOp
		case otelcol.StateRunning:
			return nil
		default:
			err = fmt.Errorf("unable to start, otelcol state is %s", state.String())
		}
	}
}

func (c *Collector) Stop() error {
	if !c.stopped {
		c.stopped = true
		c.svc.Shutdown()
	}
	<-c.appDone
	return nil
}
