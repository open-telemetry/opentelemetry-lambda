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
	"github.com/open-telemetry/opentelemetry-lambda/collector/internal/extensionAPI"
	"github.com/open-telemetry/opentelemetry-lambda/collector/internal/telemetryAPI"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/open-telemetry/opentelemetry-lambda/collector/lambdacomponents"
	"go.uber.org/zap"
)

var (
	extensionName   = filepath.Base(os.Args[0]) // extension name has to match the filename
	extensionClient = extensionAPI.NewClient(os.Getenv("AWS_LAMBDA_RUNTIME_API"))
	logger          = zap.NewExample()
)

func main() {

	logger.Debug("Launching OpenTelemetry Lambda extension", zap.String("version", Version))

	factories, _ := lambdacomponents.Components()
	collector := NewCollector(factories)
	ctx, cancel := context.WithCancel(context.Background())

	if err := collector.Start(ctx); err != nil {
		log.Fatalf("Failed to start the extension: %v", err)
	}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		s := <-sigs
		cancel()
		logger.Debug(fmt.Sprintf("Received", s))
		logger.Debug("Exiting")
	}()

	res, err := extensionClient.Register(ctx, extensionName)
	if err != nil {
		log.Fatalf("Cannot register extension: %v", err)
	}

	logger.Debug("Register ", zap.String("response :", prettyPrint(res)))

	listener := telemetryAPI.NewListener(logger)
	addr, err := listener.Start()
	if err != nil {
		log.Fatalf("Cannot start TelemetryAPI Listener: %v", err)
	}

	telemetryClient := telemetryAPI.NewClient(logger)
	_, err = telemetryClient.Subscribe(ctx, res.ExtensionID, addr)
	if err != nil {
		log.Fatalf("Cannot register TelemetryAPI client: %v", err)
	}

	// Will block until shutdown event is received or cancelled via the context.
	processEvents(ctx, collector, listener)
}

func processEvents(ctx context.Context, collector *Collector, listener *telemetryAPI.Listener) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			logger.Debug("Waiting for event...")
			res, err := extensionClient.NextEvent(ctx)
			if err != nil {
				logln("Error:", err)
				logln("Exiting")
				return
			}
			logger.Debug("Received ", zap.String("event :", prettyPrint(res)))

			err = listener.Wait(ctx, res.RequestID)
			if err != nil {
				logger.Error("problem waiting for platform.runtimeDone event", zap.Error(err), zap.String("requestID", res.RequestID))
			}

			// Exit if we receive a SHUTDOWN event
			if res.EventType == extensionAPI.Shutdown {
				collector.Stop() // TODO: handle return values
				listener.Shutdown()
				logger.Debug("Received SHUTDOWN event")
				logger.Debug("Exiting")
				return
			}
		}
	}
}
