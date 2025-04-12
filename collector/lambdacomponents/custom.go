//go:build lambdacomponents.custom

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

package lambdacomponents

import (
	custom_connector "github.com/open-telemetry/opentelemetry-lambda/collector/lambdacomponents/connector"
	custom_exporter "github.com/open-telemetry/opentelemetry-lambda/collector/lambdacomponents/exporter"
	custom_extension "github.com/open-telemetry/opentelemetry-lambda/collector/lambdacomponents/extension"
	custom_processor "github.com/open-telemetry/opentelemetry-lambda/collector/lambdacomponents/processor"
	custom_receiver "github.com/open-telemetry/opentelemetry-lambda/collector/lambdacomponents/receiver"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/otelcol"

	"go.uber.org/multierr"
)

func Components(extensionID string) (otelcol.Factories, error) {
	var errs []error

	receivers, err := makeFactoryMap(custom_receiver.Factories, extensionID)
	if err != nil {
		errs = append(errs, err)
	}

	processors, err := makeFactoryMap(custom_processor.Factories, extensionID)
	if err != nil {
		errs = append(errs, err)
	}

	exporters, err := makeFactoryMap(custom_exporter.Factories, extensionID)
	if err != nil {
		errs = append(errs, err)
	}

	extensions, err := makeFactoryMap(custom_extension.Factories, extensionID)
	if err != nil {
		errs = append(errs, err)
	}

	connectors, err := makeFactoryMap(custom_connector.Factories, extensionID)
	if err != nil {
		errs = append(errs, err)
	}

	factories := otelcol.Factories{
		Receivers:  receivers,
		Processors: processors,
		Exporters:  exporters,
		Extensions: extensions,
		Connectors: connectors,
	}

	return factories, multierr.Combine(errs...)
}

func makeFactoryMap[F component.Factory](factories []func(extensionId string) F, extensionId string) (map[component.Type]F, error) {
	preprocessedFactories := make([]F, len(factories))
	for i, f := range factories {
		preprocessedFactories[i] = f(extensionId)
	}

	return otelcol.MakeFactoryMap(preprocessedFactories...)
}
