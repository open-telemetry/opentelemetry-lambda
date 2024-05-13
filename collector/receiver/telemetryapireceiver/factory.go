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

package telemetryapireceiver // import "github.com/open-telemetry/opentelemetry-lambda/collector/receiver/telemetryapireceiver"

import (
	"context"
	"errors"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/receiver"
)

const (
	typeStr   = "telemetryapi"
	stability = component.StabilityLevelDevelopment
)

var errConfigNotTelemetryAPI = errors.New("config was not a Telemetry API receiver config")

// NewFactory creates a new receiver factory
func NewFactory(extensionID string) receiver.Factory {
	return receiver.NewFactory(
		component.MustNewType(typeStr),
		func() component.Config {
			return &Config{
				extensionID: extensionID,
			}
		},
		receiver.WithTraces(createTracesReceiver, stability))
}

func createTracesReceiver(ctx context.Context, params receiver.CreateSettings, rConf component.Config, next consumer.Traces) (receiver.Traces, error) {
	cfg, ok := rConf.(*Config)
	if !ok {
		return nil, errConfigNotTelemetryAPI
	}

	return newTelemetryAPIReceiver(cfg, next, params)
}
