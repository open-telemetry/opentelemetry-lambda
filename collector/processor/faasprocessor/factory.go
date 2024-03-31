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

package faasprocessor // import "github.com/open-telemetry/opentelemetry-lambda/collector/processor/faasprocessor"

import (
	"context"
	"errors"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/processor"
)

const (
	typeStr   = "faas"
	stability = component.StabilityLevelDevelopment
)

var errConfigNotFaas = errors.New("config was not a faas processor config")

func NewFactory() processor.Factory {
	return processor.NewFactory(
		typeStr,
		createDefaultConfig,
		processor.WithTraces(createTracesProcessor, stability),
	)
}

func createDefaultConfig() component.Config {
	return &Config{}
}

func createTracesProcessor(ctx context.Context, params processor.CreateSettings, rConf component.Config, next consumer.Traces) (processor.Traces, error) {
	_, ok := rConf.(*Config)
	if !ok {
		return nil, errConfigNotFaas
	}

	cp, err := newFaasProcessor(next, params)
	if err != nil {
		return nil, err
	}
	return cp, nil
}
