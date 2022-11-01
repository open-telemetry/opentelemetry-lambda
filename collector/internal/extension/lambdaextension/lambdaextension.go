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
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/otel/sdk/trace"
)

type lambdaExtension struct {
	config    *Config
	telemetry component.TelemetrySettings
}

// registerableTracerProvider is a tracer that supports
// the SDK methods RegisterSpanProcessor and UnregisterSpanProcessor.
//
// We use an interface instead of casting to the SDK tracer type to support tracer providers
// that extend the SDK.
type registerableTracerProvider interface {
	// RegisterSpanProcessor adds the given SpanProcessor to the list of SpanProcessors.
	// https://pkg.go.dev/go.opentelemetry.io/otel/sdk/trace#TracerProvider.RegisterSpanProcessor.
	RegisterSpanProcessor(SpanProcessor trace.SpanProcessor)

	// UnregisterSpanProcessor removes the given SpanProcessor from the list of SpanProcessors.
	// https://pkg.go.dev/go.opentelemetry.io/otel/sdk/trace#TracerProvider.UnregisterSpanProcessor.
	UnregisterSpanProcessor(SpanProcessor trace.SpanProcessor)
}

func (le *lambdaExtension) Start(_ context.Context, host component.Host) error {
	sdktracer, ok := le.telemetry.TracerProvider.(registerableTracerProvider)
	if ok {
		sdktracer.RegisterSpanProcessor(le.config.spanProcessor)
		le.telemetry.Logger.Info("Registered lambda span processor on tracer provider")
	} else {
		le.telemetry.Logger.Warn("lambda span processor registration is not available")
	}
	return nil
}

func (le *lambdaExtension) Shutdown(context.Context) error {
	sdktracer, ok := le.telemetry.TracerProvider.(registerableTracerProvider)
	if ok {
		sdktracer.UnregisterSpanProcessor(le.config.spanProcessor)
		le.telemetry.Logger.Info("Unregistered lambda span processor on tracer provider")
	} else {
		le.telemetry.Logger.Warn("lambda span processor registration is not available")
	}

	return nil
}

func newExtension(config *Config, telemetry component.TelemetrySettings) *lambdaExtension {
	return &lambdaExtension{
		config:    config,
		telemetry: telemetry,
	}
}
