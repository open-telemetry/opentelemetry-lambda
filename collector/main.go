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
	stdlog "log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/open-telemetry/opentelemetry-lambda/collector/extension"
	"github.com/open-telemetry/opentelemetry-lambda/collector/lambdacomponents"
	"go.uber.org/zap"
)

var (
	extensionName   = filepath.Base(os.Args[0]) // extension name has to match the filename
	extensionClient = extension.NewClient(os.Getenv("AWS_LAMBDA_RUNTIME_API"))
)

var log *zap.Logger

func main() {
	log, err := zap.NewProduction()
	if err != nil {
		stdlog.Fatalf("Failed to create logger: %v", err)
	}

	log.Debug("Launching OpenTelemetry Lambda extension", zap.String("version", Version))

	factories, _ := lambdacomponents.Components()
	collector := NewCollector(factories)
	ctx, cancel := context.WithCancel(context.Background())

	if err := collector.Start(ctx); err != nil {
		log.Fatal("Failed to start", zap.Error(err))
	}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		s := <-sigs
		cancel()
		log.Debug("Received signal, exiting", zap.String("signal", s.String()))
	}()

	res, err := extensionClient.Register(ctx, extensionName)
	if err != nil {
		log.Fatal("Failed to register extension", zap.Error(err))
	}

	log.Debug("Register succeeded", zap.String("extension", extensionName), zap.Any("response", res))
	// Will block until shutdown event is received or cancelled via the context.
	processEvents(ctx, collector)
}

func processEvents(ctx context.Context, collector *Collector) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			log.Debug("Waiting for event")
			res, err := extensionClient.NextEvent(ctx)
			if err != nil {
				log.Error("Extension client failed to get next event, exiting", zap.Error(err))
				return
			}

			log.Debug("Received event", zap.Any("event", res))
			// Exit if we receive a SHUTDOWN event
			if res.EventType == extension.Shutdown {
				log.Debug("Received SHUTDOWN event, exiting")
				if err := collector.Stop(); err != nil {
					log.Error("Collector did not shut down gracefully", zap.Error(err))
				}
				return
			}
			log.Debug("Event processed successfully")
		}
	}
}
