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
	"time"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config"
	"go.opentelemetry.io/collector/config/mapconverter/expandmapconverter"
	"go.opentelemetry.io/collector/config/mapprovider/envmapprovider"
	"go.opentelemetry.io/collector/config/mapprovider/filemapprovider"
	"go.opentelemetry.io/collector/config/mapprovider/yamlmapprovider"
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
	svcDone        chan struct{}
	svcErr         error
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
	providers := []config.MapProvider{filemapprovider.New(), envmapprovider.New(), yamlmapprovider.New()}
	mapProvider := make(map[string]config.MapProvider, len(providers))

	for _, provider := range providers {
		mapProvider[provider.Scheme()] = provider
	}

	cfgSet := service.ConfigProviderSettings{
		Locations:     []string{getConfig()},
		MapProviders:  mapProvider,
		MapConverters: []config.MapConverterFunc{expandmapconverter.New()},
	}
	cfgProvider, err := service.NewConfigProvider(cfgSet)

	if err != nil {
		log.Panicf("error on creating config provider: %v\n", err)
	}

	col := &Collector{
		factories:      factories,
		configProvider: cfgProvider,
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
	svc, err := service.New(params)
	if err != nil {
		return err
	}

	c.svc = svc
	c.svcDone = make(chan struct{})

	go func() {
		defer close(c.svcDone)
		c.svcErr = c.svc.Run(ctx)
	}()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("unable to start - outer context is closed with error: %w", ctx.Err())
		case <-c.svcDone:
			return fmt.Errorf("unable to start - otelcol exited prematurely with error: %w", c.svcErr)
		// Using a very short interval because there's not much
		// for the extension to do until the collector is started.
		case <-time.After(1 * time.Millisecond):
			state := c.svc.GetState()
			switch state {
			case service.Starting:
				// Keep waiting...
			case service.Running:
				return nil
			default:
				<-c.svcDone
				return fmt.Errorf("unable to start - unexpected otelcol state %d (%s): %w", state, state, c.svcErr)
			}
		}
	}
}

func (c *Collector) Stop() {
	if !c.stopped {
		c.stopped = true
		c.svc.Shutdown()
	}
	<-c.svcDone
}

func (c *Collector) Done() <-chan struct{} {
	return c.svcDone
}

func (c *Collector) Err() error {
	return c.svcErr
}
