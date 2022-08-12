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
	extensionName   = filepath.Base(os.Args[0]) // extension name has to match the filename
	extensionClient = extension.NewClient(os.Getenv("AWS_LAMBDA_RUNTIME_API"))
	logger          *zap.SugaredLogger
)

func main() {
	configureLogger()
	defer logger.Sync()

	logger.Debugw("Launching OpenTelemetry Lambda extension", "version", Version)

	factories, _ := lambdacomponents.Components()
	collector := NewCollector(factories)
	ctx, cancel := context.WithCancel(context.Background())

	if err := collector.Start(ctx); err != nil {
		logger.Fatalf("Failed to start the extension: %v", err)
	}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		s := <-sigs
		cancel()
		logger.Debugf("Received: %s", s.String())
		logger.Debug("Exiting")
	}()

	res, err := extensionClient.Register(ctx, extensionName)
	if err != nil {
		logger.Fatalf("Cannot register extension: %v", err)
	}

	logger.Debugw("Register", "response", res)
	// Will block until shutdown event is received or cancelled via the context.
	processEvents(ctx, collector)
}

func configureLogger() {
	atom := zap.NewAtomicLevel()

	level, err := zapcore.ParseLevel(os.Getenv("OPENTELEMETRY_COLLECTOR_LOG_LEVEL"))
	if err != nil {
		level = zap.DebugLevel
	}

	atom.SetLevel(level)

	l := zap.New(zapcore.NewCore(zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()), os.Stdout, atom))
	zap.ReplaceGlobals(l)

	logger = l.Sugar()
}

func processEvents(ctx context.Context, collector *Collector) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			logger.Debug("Waiting for event...")
			res, err := extensionClient.NextEvent(ctx)
			if err != nil {
				logger.Errorf("[%s] Error: %v", extensionName, err)
				logger.Debugf("[%s] Exiting", extensionName)
				return
			}

			logger.Debugw("Received", "event", res)
			// Exit if we receive a SHUTDOWN event
			if res.EventType == extension.Shutdown {
				collector.Stop() // TODO: handle return values
				logger.Debug("Received SHUTDOWN event")
				logger.Debug("Exiting")
				return
			}
		}
	}
}
