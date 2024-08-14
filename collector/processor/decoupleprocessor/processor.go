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
	"sync"

	"github.com/open-telemetry/opentelemetry-lambda/collector/lambdalifecycle"
	"go.opentelemetry.io/collector/client"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/processor/processorhelper"

	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/pdata/ptrace"
	"go.opentelemetry.io/collector/processor"
	"go.uber.org/zap"
)

var (
	incorrectDataTypeError   = errors.New("incorrect data type")
	noLifecycleNotifierError = errors.New("no lifecycle notifier set")
)

type decoupleConsumer interface {
	consume(context.Context, any) error
}

type contextualData struct {
	info client.Info
	data any
}

type decoupleProcessor struct {
	logger   *zap.Logger
	consumer decoupleConsumer

	data chan contextualData

	wg sync.WaitGroup
}

func (p *decoupleProcessor) queueData(ctx context.Context, data any) {
	p.data <- contextualData{
		info: client.FromContext(ctx),
		data: data,
	}
}

func (p *decoupleProcessor) processTraces(ctx context.Context, td ptrace.Traces) (ptrace.Traces, error) {
	p.queueData(ctx, &td)
	return td, processorhelper.ErrSkipProcessingData
}

func (p *decoupleProcessor) processMetrics(ctx context.Context, md pmetric.Metrics) (pmetric.Metrics, error) {
	p.queueData(ctx, &md)
	return md, processorhelper.ErrSkipProcessingData
}

func (p *decoupleProcessor) processLogs(ctx context.Context, ld plog.Logs) (plog.Logs, error) {
	p.queueData(ctx, &ld)
	return ld, processorhelper.ErrSkipProcessingData
}

func (p *decoupleProcessor) startForwardingData() {
	p.wg.Add(1)
	go func() {
		defer p.wg.Done()
		p.logger.Info("started forwarding data")
		for {
			d := <-p.data
			if d.data == nil {
				break
			}
			if err := p.consumer.consume(client.NewContext(context.Background(), d.info), d.data); err != nil {
				p.logger.Error("next consumer failed", zap.Error(err))
			}
		}
		p.logger.Info("stopped forwarding data")
	}()
}

func (p *decoupleProcessor) stopForwardingData() {
	p.data <- contextualData{}
	p.wg.Wait()
}

func (p *decoupleProcessor) shutdown(ctx context.Context) error {
	p.stopForwardingData()
	return nil
}

func (p *decoupleProcessor) FunctionInvoked() {
	p.startForwardingData()
}

func (p *decoupleProcessor) FunctionFinished() {
	// Stop forwarding data to ensure that we don't have issues with network interruptions if the environment is frozen.
	p.stopForwardingData()
}

func (p *decoupleProcessor) EnvironmentShutdown() {
	// Start the forwarder to ensure any traces left in the pipeline can be sent when the collector is shutdown.
	p.startForwardingData()
}

func newDecoupleProcessor(
	cfg *Config,
	consumer decoupleConsumer,
	set processor.Settings,
) (*decoupleProcessor, error) {
	dp := &decoupleProcessor{
		consumer: consumer,
		logger:   set.Logger,
		data:     make(chan contextualData, cfg.MaxQueueSize),
	}
	if notifier := lambdalifecycle.GetNotifier(); notifier == nil {
		return nil, noLifecycleNotifierError
	} else {
		notifier.AddListener(dp)
	}
	return dp, nil
}

type decoupleTraceConsumer struct {
	nextConsumer consumer.Traces
}

func (tc *decoupleTraceConsumer) consume(ctx context.Context, data any) error {
	if td, ok := data.(*ptrace.Traces); ok {
		return tc.nextConsumer.ConsumeTraces(ctx, *td)
	} else {
		return incorrectDataTypeError
	}
}

func newDecoupleTracesProcessor(cfg *Config,
	next consumer.Traces,
	set processor.Settings,
) (*decoupleProcessor, error) {
	return newDecoupleProcessor(cfg, &decoupleTraceConsumer{nextConsumer: next}, set)
}

type decoupleMetricsConsumer struct {
	nextConsumer consumer.Metrics
}

func (tc *decoupleMetricsConsumer) consume(ctx context.Context, data any) error {
	if md, ok := data.(*pmetric.Metrics); ok {
		return tc.nextConsumer.ConsumeMetrics(ctx, *md)
	} else {
		return incorrectDataTypeError
	}
}

func newDecoupleMetricsProcessor(cfg *Config,
	next consumer.Metrics,
	set processor.Settings,
) (*decoupleProcessor, error) {
	return newDecoupleProcessor(cfg, &decoupleMetricsConsumer{nextConsumer: next}, set)
}

type decoupleLogsConsumer struct {
	nextConsumer consumer.Logs
}

func (tc *decoupleLogsConsumer) consume(ctx context.Context, data any) error {
	if ld, ok := data.(*plog.Logs); ok {
		return tc.nextConsumer.ConsumeLogs(ctx, *ld)
	} else {
		return incorrectDataTypeError
	}
}

func newDecoupleLogsProcessor(cfg *Config,
	next consumer.Logs,
	set processor.Settings,
) (*decoupleProcessor, error) {
	return newDecoupleProcessor(cfg, &decoupleLogsConsumer{nextConsumer: next}, set)
}
