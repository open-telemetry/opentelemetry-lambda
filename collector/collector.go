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

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/service"
)

var (
	// Version variable will be replaced at link time after `make` has been run.
	Version = "latest"

	// GitHash variable will be replaced at link time after `make` has been run.
	GitHash = "<NOT PROPERLY GENERATED>"
)

// Collector implements the OtelcolRunner interfaces running a single otelcol as a go routine within the
// same process as the test executor.
type Collector struct {
	factories      component.Factories
	configProvider service.ConfigProvider
	svc            *service.Collector
	appDone        chan struct{}
	stopped        bool
}

func getConfig() string {
	val, ex := os.LookupEnv("OPENTELEMETRY_COLLECTOR_CONFIG_FILE")
	if !ex {
		return "/opt/collector-config/config.yaml"
	}
	log.Printf("Using config file at path %v", val)
	return val
}

func NewCollector(factories component.Factories) *Collector {
	col := &Collector{
		factories:      factories,
		configProvider: service.NewDefaultConfigProvider([]string{getConfig()}, nil),
	}
	return col
}

func (c *Collector) Start(ctx context.Context) error {
	params := service.CollectorSettings{
		BuildInfo: component.BuildInfo{
			Command:     "otelcol",
			Description: "Lambda Collector",
			Version:     Version,
		},
		ConfigProvider: c.configProvider,
		Factories:      c.factories,
	}
	var err error
	c.svc, err = service.New(params)
	if err != nil {
		return err
	}

	c.appDone = make(chan struct{})
	runErr := make(chan error, 1)
	go func() {
		defer close(c.appDone)
		err := c.svc.Run(ctx)
		runErr <- err
	}()

	rErr := <-runErr
	if rErr != nil {
		return rErr
	}
	close(runErr)

	state := c.svc.GetState()

	switch state {
	case service.Starting:
		// NoOp
	case service.Running:
		return err
	default:
		err = fmt.Errorf("unable to start, otelcol state is %d", state)
	}

	return err
}

func (c *Collector) Stop() error {
	if !c.stopped {
		c.stopped = true
		c.svc.Shutdown()
	}
	<-c.appDone
	return nil
}
