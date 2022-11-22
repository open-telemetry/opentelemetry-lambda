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

package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/open-telemetry/opentelemetry-lambda/collector/extension"
	"github.com/open-telemetry/opentelemetry-lambda/collector/lambdacomponents"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	extensionName = filepath.Base(os.Args[0]) // extension name has to match the filename
)

func main() {
	logger := initLogger()
	logger.Info("Launching OpenTelemetry Lambda extension", zap.String("version", Version))

	ctx, lm := newLifecycleManager(context.Background(), logger)

	// Will block until shutdown event is received or cancelled via the context.
	lm.processEvents(ctx)
}

type lifecycleManager struct {
	logger          *zap.Logger
	collector       *Collector
	extensionClient *extension.Client
}

func newLifecycleManager(ctx context.Context, logger *zap.Logger) (context.Context, *lifecycleManager) {
	ctx, cancel := context.WithCancel(ctx)

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		s := <-sigs
		cancel()
		logger.Info("received signal", zap.String("signal", s.String()))
	}()

	extensionClient := extension.NewClient(logger, os.Getenv("AWS_LAMBDA_RUNTIME_API"))
	res, err := extensionClient.Register(ctx, extensionName)
	if err != nil {
		log.Fatalf("Cannot register extension: %v", err)
	}

	logger.Debug("register extension", zap.Any("response :", res))

	factories, _ := lambdacomponents.Components()
	collector := NewCollector(logger, factories)

	if err = collector.Start(ctx); err != nil {
		logger.Fatal("Failed to start the extension", zap.Error(err))
		extensionClient.InitError(ctx, fmt.Sprintf("failed to start the collector: %v", err))
	}

	return ctx, &lifecycleManager{
		logger:          logger.Named("lifecycleManager"),
		collector:       collector,
		extensionClient: extensionClient,
	}
}

func (lm *lifecycleManager) processEvents(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			lm.logger.Debug("Waiting for event...")
			res, err := lm.extensionClient.NextEvent(ctx)
			if err != nil {
				lm.logger.Warn("error waiting for extension event", zap.Error(err))
				lm.extensionClient.ExitError(ctx, fmt.Sprintf("error waiting for extension event: %v", err))
				return
			}

			lm.logger.Debug("Received ", zap.Any("event :", res))
			// Exit if we receive a SHUTDOWN event
			if res.EventType == extension.Shutdown {
				lm.logger.Info("Received SHUTDOWN event")
				err = lm.collector.Stop()
				if err != nil {
					lm.extensionClient.ExitError(ctx, fmt.Sprintf("error stopping collector: %v", err))
				}
				return
			}
		}
	}
}

func initLogger() *zap.Logger {
	lvl := zap.NewAtomicLevelAt(zapcore.InfoLevel)

	envLvl := os.Getenv("OPENTELEMETRY_EXTENSION_LOG_LEVEL")
	userLvl, err := zap.ParseAtomicLevel(envLvl)
	if err == nil {
		lvl = userLvl
	}

	l := zap.New(zapcore.NewCore(zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()), os.Stdout, lvl))

	if err != nil && envLvl != "" {
		l.Warn("unable to parse log level from environment", zap.Error(err))
	}

	return l
}
