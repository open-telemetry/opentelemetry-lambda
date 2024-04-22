// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package decoupleprocessor // import "github.com/open-telemetry/opentelemetry-lambda/collector/processor/decoupleprocessor"

import (
	"context"
	"errors"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/processor"
	"go.opentelemetry.io/collector/processor/processorhelper"
)

const (
	typeStr   = "decouple"
	stability = component.StabilityLevelDevelopment
)

var (
	errConfigNotDecouple  = errors.New("config was not a decouple processor config")
	processorCapabilities = consumer.Capabilities{MutatesData: false}
)

func NewFactory() processor.Factory {
	return processor.NewFactory(
		component.MustNewType(typeStr),
		createDefaultConfig,
		processor.WithTraces(createTracesProcessor, stability),
		processor.WithMetrics(createMetricsProcessor, stability),
		processor.WithLogs(createLogsProcessor, stability),
	)
}

func createDefaultConfig() component.Config {
	return &Config{
		MaxQueueSize: 200,
	}
}

func createTracesProcessor(ctx context.Context, params processor.CreateSettings, rConf component.Config, next consumer.Traces) (processor.Traces, error) {
	cfg, ok := rConf.(*Config)
	if !ok {
		return nil, errConfigNotDecouple
	}

	dp, err := newDecoupleTracesProcessor(cfg, next, params)
	if err != nil {
		return nil, err
	}
	return processorhelper.NewTracesProcessor(
		ctx,
		params,
		cfg,
		next,
		dp.processTraces,
		processorhelper.WithCapabilities(processorCapabilities),
		processorhelper.WithShutdown(dp.shutdown),
	)
}

func createMetricsProcessor(ctx context.Context, params processor.CreateSettings, rConf component.Config, next consumer.Metrics) (processor.Metrics, error) {
	cfg, ok := rConf.(*Config)
	if !ok {
		return nil, errConfigNotDecouple
	}
	dp, err := newDecoupleMetricsProcessor(cfg, next, params)
	if err != nil {
		return nil, err
	}

	return processorhelper.NewMetricsProcessor(
		ctx,
		params,
		cfg,
		next,
		dp.processMetrics,
		processorhelper.WithCapabilities(processorCapabilities),
		processorhelper.WithShutdown(dp.shutdown),
	)
}

func createLogsProcessor(ctx context.Context, params processor.CreateSettings, rConf component.Config, next consumer.Logs) (processor.Logs, error) {
	cfg, ok := rConf.(*Config)
	if !ok {
		return nil, errConfigNotDecouple
	}
	dp, err := newDecoupleLogsProcessor(cfg, next, params)
	if err != nil {
		return nil, err
	}

	return processorhelper.NewLogsProcessor(
		ctx,
		params,
		cfg,
		next,
		dp.processLogs,
		processorhelper.WithCapabilities(processorCapabilities),
		processorhelper.WithShutdown(dp.shutdown),
	)
}
