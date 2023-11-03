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
	"github.com/open-telemetry/opentelemetry-collector-contrib/exporter/prometheusremotewriteexporter"
	"github.com/open-telemetry/opentelemetry-collector-contrib/extension/sigv4authextension"
	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/attributesprocessor"
	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/filterprocessor"
	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/probabilisticsamplerprocessor"
	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/resourcedetectionprocessor"
	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/resourceprocessor"
	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/spanprocessor"
	"github.com/open-telemetry/opentelemetry-lambda/collector/processor/decoupleprocessor"
	"github.com/solarwinds/opentelemetry-collector-contrib/extension/solarwindsapmsettingsextension"
	"go.opentelemetry.io/collector/exporter"
	"go.opentelemetry.io/collector/exporter/loggingexporter"
	"go.opentelemetry.io/collector/exporter/otlpexporter"
	"go.opentelemetry.io/collector/exporter/otlphttpexporter"
	"go.opentelemetry.io/collector/extension"
	"go.opentelemetry.io/collector/otelcol"
	"go.opentelemetry.io/collector/processor"
	"go.opentelemetry.io/collector/processor/batchprocessor"
	"go.opentelemetry.io/collector/processor/memorylimiterprocessor"
	"go.opentelemetry.io/collector/receiver"
	"go.opentelemetry.io/collector/receiver/otlpreceiver"
	"go.uber.org/multierr"

	"github.com/open-telemetry/opentelemetry-lambda/collector/processor/coldstartprocessor"
	"github.com/open-telemetry/opentelemetry-lambda/collector/receiver/telemetryapireceiver"
)

func Components(extensionID string) (otelcol.Factories, error) {
	var errs []error

	receivers, err := receiver.MakeFactoryMap(
		otlpreceiver.NewFactory(),
		telemetryapireceiver.NewFactory(extensionID),
	)
	if err != nil {
		errs = append(errs, err)
	}

	exporters, err := exporter.MakeFactoryMap(
		loggingexporter.NewFactory(),
		otlpexporter.NewFactory(),
		otlphttpexporter.NewFactory(),
		prometheusremotewriteexporter.NewFactory(),
	)
	if err != nil {
		errs = append(errs, err)
	}

	processors, err := processor.MakeFactoryMap(
		attributesprocessor.NewFactory(),
		filterprocessor.NewFactory(),
		memorylimiterprocessor.NewFactory(),
		probabilisticsamplerprocessor.NewFactory(),
		resourceprocessor.NewFactory(),
		spanprocessor.NewFactory(),
		coldstartprocessor.NewFactory(),
		decoupleprocessor.NewFactory(),
		batchprocessor.NewFactory(),
		resourcedetectionprocessor.NewFactory(),
	)
	if err != nil {
		errs = append(errs, err)
	}

	extensions, err := extension.MakeFactoryMap(
		sigv4authextension.NewFactory(),
		solarwindsapmsettingsextension.NewFactory(),
	)
	if err != nil {
		errs = append(errs, err)
	}

	factories := otelcol.Factories{
		Receivers:  receivers,
		Exporters:  exporters,
		Processors: processors,
		Extensions: extensions,
	}

	return factories, multierr.Combine(errs...)
}
