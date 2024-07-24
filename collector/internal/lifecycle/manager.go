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

package lifecycle

import (
	"context"
	"fmt"
	"github.com/open-telemetry/opentelemetry-lambda/collector/lambdalifecycle"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"

	"go.uber.org/multierr"
	"go.uber.org/zap"

	"github.com/open-telemetry/opentelemetry-lambda/collector/internal/collector"
	"github.com/open-telemetry/opentelemetry-lambda/collector/internal/extensionapi"
	"github.com/open-telemetry/opentelemetry-lambda/collector/internal/telemetryapi"
	"github.com/open-telemetry/opentelemetry-lambda/collector/lambdacomponents"
)

var (
	extensionName = filepath.Base(os.Args[0]) // extension name has to match the filename
)

type collectorWrapper interface {
	Start(ctx context.Context) error
	Stop() error
}

type manager struct {
	logger             *zap.Logger
	collector          collectorWrapper
	extensionClient    *extensionapi.Client
	listener           *telemetryapi.Listener
	wg                 sync.WaitGroup
	lifecycleListeners []lambdalifecycle.Listener
}

func NewManager(ctx context.Context, logger *zap.Logger, version string) (context.Context, *manager) {
	ctx, cancel := context.WithCancel(ctx)

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		s := <-sigs
		cancel()
		logger.Info("received signal", zap.String("signal", s.String()))
	}()

	extensionClient := extensionapi.NewClient(logger, os.Getenv("AWS_LAMBDA_RUNTIME_API"))
	res, err := extensionClient.Register(ctx, extensionName)
	if err != nil {
		logger.Fatal("Cannot register extension", zap.Error(err))
	}

	listener := telemetryapi.NewListener(logger)
	addr, err := listener.Start()
	if err != nil {
		logger.Fatal("Cannot start Telemetry API Listener", zap.Error(err))
	}

	telemetryClient := telemetryapi.NewClient(logger)
	_, err = telemetryClient.Subscribe(ctx, []telemetryapi.EventType{telemetryapi.Platform}, res.ExtensionID, addr)
	if err != nil {
		logger.Fatal("Cannot register Telemetry API client", zap.Error(err))
	}

	lm := &manager{
		logger:          logger.Named("lifecycle.manager"),
		extensionClient: extensionClient,
		listener:        listener,
	}

	factories, _ := lambdacomponents.Components(res.ExtensionID)
	lm.collector = collector.NewCollector(logger, factories, version)

	return ctx, lm
}

func (lm *manager) Run(ctx context.Context) error {
	if err := lm.collector.Start(ctx); err != nil {
		lm.logger.Warn("Failed to start the extension", zap.Error(err))
		if _, initErr := lm.extensionClient.InitError(ctx, fmt.Sprintf("failed to start the collector: %v", err)); initErr != nil {
			return multierr.Combine(err, initErr)
		}
		return err
	}

	lm.wg.Add(1)
	go func() {
		if err := lm.processEvents(ctx); err != nil {
			lm.logger.Warn("Failed to process events", zap.Error(err))
		}
	}()

	lm.wg.Wait()
	return nil
}

func (lm *manager) processEvents(ctx context.Context) error {
	defer lm.wg.Done()

	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			lm.logger.Debug("Waiting for event...")
			res, err := lm.extensionClient.NextEvent(ctx)
			if err != nil {
				lm.logger.Warn("error waiting for extension event", zap.Error(err))
				if _, exitErr := lm.extensionClient.ExitError(ctx, fmt.Sprintf("error waiting for extension event: %v", err)); exitErr != nil {
					return multierr.Combine(err, exitErr)
				}
				return err
			}

			lm.logger.Debug("Received ", zap.Any("event :", res))
			// Exit if we receive a SHUTDOWN event
			if res.EventType == extensionapi.Shutdown {
				lm.logger.Info("Received SHUTDOWN event")
				lm.notifyEnvironmentShutdown()
				lm.listener.Shutdown()
				err = lm.collector.Stop()
				if err != nil {
					if _, exitErr := lm.extensionClient.ExitError(ctx, fmt.Sprintf("error stopping collector: %v", err)); exitErr != nil {
						return multierr.Combine(err, exitErr)
					}
				}
				return err
			}

			lm.notifyFunctionInvoked()

			err = lm.listener.Wait(ctx, res.RequestID)
			if err != nil {
				lm.logger.Error("problem waiting for platform.runtimeDone event", zap.Error(err), zap.String("requestID", res.RequestID))
			}

			// Check other components are ready before allowing the freezing of the environment.
			lm.notifyFunctionFinished()
		}
	}
}

func (lm *manager) notifyFunctionInvoked() {
	for _, listener := range lm.lifecycleListeners {
		listener.FunctionInvoked()
	}
}

func (lm *manager) notifyFunctionFinished() {
	for _, listener := range lm.lifecycleListeners {
		listener.FunctionFinished()
	}
}

func (lm *manager) notifyEnvironmentShutdown() {
	for _, listener := range lm.lifecycleListeners {
		listener.EnvironmentShutdown()
	}
}

func (lm *manager) AddListener(listener lambdalifecycle.Listener) {
	lm.lifecycleListeners = append(lm.lifecycleListeners, listener)
}
