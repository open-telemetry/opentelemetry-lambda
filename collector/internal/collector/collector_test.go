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

package collector

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/exporter/exportertest"
	"go.opentelemetry.io/collector/otelcol"
	"go.opentelemetry.io/collector/receiver/receivertest"
	"go.opentelemetry.io/collector/service/telemetry/otelconftelemetry"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

func TestCollectorConfigLogLevelIsOverridden(t *testing.T) {
	t.Setenv("OPENTELEMETRY_COLLECTOR_CONFIG_URI", "file:testdata/config-error-level.yaml")

	receivers, err := otelcol.MakeFactoryMap(receivertest.NewNopFactory())
	require.NoError(t, err)
	exporters, err := otelcol.MakeFactoryMap(exportertest.NewNopFactory())
	require.NoError(t, err)

	factories := otelcol.Factories{
		Receivers: receivers,
		Exporters: exporters,
		Telemetry: otelconftelemetry.NewFactory(),
	}

	// Use a nop logger so extension logs don't end up in our observer
	collector := NewCollector(zap.NewNop(), factories, "test")
	// Replace collector logger with an observed core at INFO level.
	collectorObservedCore, collectorLogs := observer.New(zapcore.InfoLevel)
	collector.logger = zap.New(collectorObservedCore)

	ctx := context.Background()
	err = collector.Start(ctx)
	require.NoError(t, err)

	err = collector.Stop()
	require.NoError(t, err)

	// If the collector config log level is respected, there are no INFO logs
	infoLogs := collectorLogs.FilterLevelExact(zapcore.InfoLevel).All()
	assert.Empty(t, infoLogs,
		"INFO logs from the collector should be suppressed when config sets level: error")
}
