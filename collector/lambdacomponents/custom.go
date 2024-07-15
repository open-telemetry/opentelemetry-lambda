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
	_ "github.com/open-telemetry/opentelemetry-lambda/collector/lambdacomponents/connector"
	_ "github.com/open-telemetry/opentelemetry-lambda/collector/lambdacomponents/exporter"
	_ "github.com/open-telemetry/opentelemetry-lambda/collector/lambdacomponents/extension"
	_ "github.com/open-telemetry/opentelemetry-lambda/collector/lambdacomponents/processor"
	_ "github.com/open-telemetry/opentelemetry-lambda/collector/lambdacomponents/receiver"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/connector"
	"go.opentelemetry.io/collector/exporter"
	"go.opentelemetry.io/collector/extension"
	"go.opentelemetry.io/collector/otelcol"
	"go.opentelemetry.io/collector/processor"
	"go.opentelemetry.io/collector/receiver"

	"go.uber.org/multierr"
)

type factory[T any] interface {
	Get(extensionId string) T
}

type FactoryFn[T any] func() T

func (fn FactoryFn[T]) Get(extensionId string) T {
	return fn()
}

type FactoryWithExtensionIdFn[T any] func(extensionId string) T

func (fn FactoryWithExtensionIdFn[T]) Get(extensionId string) T {
	return fn(extensionId)
}

var (
	receiverFactories  []factory[receiver.Factory]
	processorFactories []factory[processor.Factory]
	exporterFactories  []factory[exporter.Factory]
	extensionFactories []factory[extension.Factory]
	connectorFactories []factory[connector.Factory]
)

func AddReceiverFactory(f factory[receiver.Factory]) {
	receiverFactories = append(receiverFactories, f)
}

func AddProcessorFactory(f factory[processor.Factory]) {
	processorFactories = append(processorFactories, f)
}

func AddExporterFactory(f factory[exporter.Factory]) {
	exporterFactories = append(exporterFactories, f)
}

func AddExtensionFactory(f factory[extension.Factory]) {
	extensionFactories = append(extensionFactories, f)
}

func AddConnectorFactory(f factory[connector.Factory]) {
	connectorFactories = append(connectorFactories, f)
}

func Components(extensionID string) (otelcol.Factories, error) {
	var errs []error

	receivers, err := makeFactoryMap(receiverFactories, receiver.MakeFactoryMap, extensionID)
	if err != nil {
		errs = append(errs, err)
	}

	processors, err := makeFactoryMap(processorFactories, processor.MakeFactoryMap, extensionID)
	if err != nil {
		errs = append(errs, err)
	}

	exporters, err := makeFactoryMap(exporterFactories, exporter.MakeFactoryMap, extensionID)
	if err != nil {
		errs = append(errs, err)
	}

	extensions, err := makeFactoryMap(extensionFactories, extension.MakeFactoryMap, extensionID)
	if err != nil {
		errs = append(errs, err)
	}

	connectors, err := makeFactoryMap(connectorFactories, connector.MakeFactoryMap, extensionID)
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

func makeFactoryMap[F any](factories []factory[F], fn func(...F) (map[component.Type]F, error), extensionId string) (map[component.Type]F, error) {
	preprocessedFactories := make([]F, len(factories))
	for i, f := range factories {
		preprocessedFactories[i] = f.Get(extensionId)
	}

	return fn(preprocessedFactories...)
}
