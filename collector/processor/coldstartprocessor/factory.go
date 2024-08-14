// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package coldstartprocessor // import "github.com/open-telemetry/opentelemetry-lambda/collector/processor/coldstartprocessor"

import (
	"context"
	"errors"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/processor"
	"go.opentelemetry.io/collector/processor/processorhelper"
)

const (
	typeStr   = "coldstart"
	stability = component.StabilityLevelDevelopment
)

var (
	errConfigNotColdstart = errors.New("config was not a Coldstart processor config")
	processorCapabilities = consumer.Capabilities{MutatesData: true}
)

func NewFactory() processor.Factory {
	return processor.NewFactory(
		component.MustNewType(typeStr),
		createDefaultConfig,
		processor.WithTraces(createTracesProcessor, stability),
	)
}

func createDefaultConfig() component.Config {
	return &Config{}
}

func createTracesProcessor(ctx context.Context, params processor.Settings, rConf component.Config, next consumer.Traces) (processor.Traces, error) {
	cfg, ok := rConf.(*Config)
	if !ok {
		return nil, errConfigNotColdstart
	}

	cp, err := newColdstartProcessor(cfg, next, params)
	if err != nil {
		return nil, err
	}
	return processorhelper.NewTracesProcessor(
		ctx,
		params,
		cfg,
		next,
		cp.processTraces,
		processorhelper.WithCapabilities(processorCapabilities),
	)

}
