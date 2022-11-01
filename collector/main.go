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
	"github.com/open-telemetry/opentelemetry-lambda/collector/internal/extension/lambdaextension"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"go.uber.org/zap"

	"github.com/open-telemetry/opentelemetry-lambda/collector/extension"
	"github.com/open-telemetry/opentelemetry-lambda/collector/lambdacomponents"
)

var (
	extensionName   = filepath.Base(os.Args[0]) // extension name has to match the filename
	extensionClient = extension.NewClient(os.Getenv("AWS_LAMBDA_RUNTIME_API"))
	logger          = zap.NewExample()
	sp              = newSpanProcessor()
)

func main() {

	logger.Debug("Launching OpenTelemetry Lambda extension", zap.String("version", Version))

	factories, _ := lambdacomponents.Components()
	lambdaFactory := lambdaextension.NewFactory(sp)
	factories.Extensions[lambdaFactory.Type()] = lambdaFactory
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
	// Will block until shutdown event is received or cancelled via the context.
	processEvents(ctx, collector)
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
				logln("Error:", err)
				logln("Exiting")
				return
			}

			logger.Debug("Received ", zap.String("event :", prettyPrint(res)))
			// Exit if we receive a SHUTDOWN event
			if res.EventType == extension.Shutdown {
				collector.Stop() // TODO: handle return values
				logger.Debug("Received SHUTDOWN event")
				logger.Debug("Exiting")
				return
			}

			select {
			case <-sp.waitCh:
			case <-time.After(1000 * time.Millisecond):
			}

			for c := sp.activeSpanCount(); c > 0; c = sp.activeSpanCount() {
				logger.Info("Waiting for quiescence", zap.Int("active_spans", c))
				time.Sleep(1 * time.Millisecond)
			}
		}
	}
}
