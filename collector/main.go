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
)

var (
	extensionName   = filepath.Base(os.Args[0]) // extension name has to match the filename
	extensionClient = extension.NewClient(os.Getenv("AWS_LAMBDA_RUNTIME_API"))
)

func main() {
	logln("Launching Opentelemetry Lambda extension, version: ", Version)

	factories, _ := lambdacomponents.Components()
	collector := NewInProcessCollector(factories)
	collector.prepareConfig()
	collector.start()

	ctx, cancel := context.WithCancel(context.Background())

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		s := <-sigs
		cancel()
		logln("Received", s)
		logln("Exiting")
	}()

	res, err := extensionClient.Register(ctx, extensionName)
	if err != nil {
		panic(err)
	}
	logln("Register response:", prettyPrint(res))

	// Will block until shutdown event is received or cancelled via the context.
	processEvents(ctx, collector)
}

func processEvents(ctx context.Context, collector *InProcessCollector) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			logln("Waiting for event...")
			res, err := extensionClient.NextEvent(ctx)
			if err != nil {
				logln("Error:", err)
				logln("Exiting")
				return
			}

			logln("Received event:", prettyPrint(res))
			// Exit if we receive a SHUTDOWN event
			if res.EventType == extension.Shutdown {
				collector.stop() // TODO: handle return values
				logln("Received SHUTDOWN event")
				logln("Exiting")
				return
			}
		}
	}
}
