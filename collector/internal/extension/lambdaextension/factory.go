// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package lambdaextension // import "github.com/open-telemetry/opentelemetry-lambda/collector/internal/extension/lambdaextension"

import (
	"context"
	"go.opentelemetry.io/otel/sdk/trace"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config"
)

const (
	// The value of extension "type" in configuration.
	typeStr = "lambda"
)

// NewFactory creates a factory for lambda extension.
func NewFactory(sp trace.SpanProcessor) component.ExtensionFactory {
	return component.NewExtensionFactory(typeStr, createDefaultConfig(sp), createExtension, component.StabilityLevelInDevelopment)
}

func createDefaultConfig(sp trace.SpanProcessor) func() config.Extension {
	return func() config.Extension {
		return &Config{
			ExtensionSettings: config.NewExtensionSettings(config.NewComponentID(typeStr)),
			spanProcessor:     sp,
		}
	}
}

// createExtension creates the extension based on this config.
func createExtension(_ context.Context, set component.ExtensionCreateSettings, cfg config.Extension) (component.Extension, error) {
	return newExtension(cfg.(*Config), set.TelemetrySettings), nil
}
