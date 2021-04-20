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
	"bytes"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config"
	"go.opentelemetry.io/collector/config/configmodels"
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
	factories component.Factories
	config    *configmodels.Config
	svc       *service.Application
	appDone   chan struct{}
	stopped   bool
}

var configFile = getConfig()

func getConfig() string {
	val, ex := os.LookupEnv("OPENTELEMETRY_COLLECTOR_CONFIG_FILE")
	if !ex {
		return "/opt/collector-config/config.yaml"
	}
	return val
}

// NewCollector creates a new Lambda collector using the supplied component factories.
func NewCollector(factories component.Factories) *Collector {
	col := &Collector{
		factories: factories,
	}
	col.prepareConfig()
	return col
}

func envLoaderConfigFactory(v *viper.Viper, factories component.Factories) (*configmodels.Config, error) {
	if configContent, ok := os.LookupEnv("OPENTELEMETRY_COLLECTOR_CONFIG_CONTENT"); ok {
		logln("Reading config from environment: ", configContent)
		configContent = strings.Replace(configContent, "\\n", "\n", -1)
		var configBytes = []byte(configContent)
		err := v.ReadConfig(bytes.NewBuffer(configBytes))
		if err != nil {
			return nil, fmt.Errorf("error loading config %v", err)
		}
		return config.Load(v, factories)
	}

	logln("Reading config from file: ", configFile)
	file := configFile
	if file == "" {
		return nil, errors.New("config file not specified")
	}
	v.SetConfigFile(file)
	err := v.ReadInConfig()
	if err != nil {
		return nil, fmt.Errorf("error loading config file %q: %v", file, err)
	}
	return config.Load(v, factories)
}

func (c *Collector) prepareConfig() (err error) {
	v := config.NewViper()
	v.SetConfigType("yaml")
	cfg, err := envLoaderConfigFactory(v, c.factories)
	if err != nil {
		return err
	}
	c.config = cfg
	return err
}

func (c *Collector) Start() error {
	params := service.Parameters{
		ApplicationStartInfo: component.ApplicationStartInfo{
			ExeName:  "otelcol",
			LongName: "Lambda Collector",
			Version:  Version,
			GitHash:  GitHash,
		},
		ConfigFactory: func(v *viper.Viper, cmd *cobra.Command, factories component.Factories) (*configmodels.Config, error) {
			return c.config, nil
		},
		Factories: c.factories,
	}
	var err error
	c.svc, err = service.New(params)
	if err != nil {
		return err
	}
	c.svc.Command().SetArgs([]string{"--metrics-level=NONE"})

	c.appDone = make(chan struct{})
	go func() {
		defer close(c.appDone)
		appErr := c.svc.Run()
		if appErr != nil {
			err = appErr
		}
	}()

	for state := range c.svc.GetStateChannel() {
		switch state {
		case service.Starting:
			// NoOp
		case service.Running:
			return err
		default:
			err = fmt.Errorf("unable to start, otelcol state is %d", state)
		}
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
