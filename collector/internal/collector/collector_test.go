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
	"encoding/base64"
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

func TestCollectorConfigLogLevelSuppressesCollectorInfoLogs(t *testing.T) {
	t.Setenv("OPENTELEMETRY_COLLECTOR_CONFIG_URI", "file:testdata/config-error-level.yaml")

	collectorLogs := &observer.ObservedLogs{}
	collector := NewCollector(zap.NewNop(), testFactories(t), "test")
	collector.coreFunc = func(levelEnabler zapcore.LevelEnabler) zapcore.Core {
		var collectorObservedCore zapcore.Core
		collectorObservedCore, collectorLogs = observer.New(levelEnabler)
		return collectorObservedCore
	}

	ctx := context.Background()
	err := collector.Start(ctx)
	require.NoError(t, err)

	err = collector.Stop()
	require.NoError(t, err)

	// If the collector config log level is respected, there are no INFO logs
	infoLogs := collectorLogs.FilterLevelExact(zapcore.InfoLevel).All()
	assert.Empty(t, infoLogs,
		"INFO logs from the collector should be suppressed when config sets level: error")
}

func TestExtensionLogLevelDoesNotSuppressCollectorLogs(t *testing.T) {
	t.Setenv("OPENTELEMETRY_COLLECTOR_CONFIG_URI", "file:testdata/config-info-level.yaml")

	extensionObservedCore, extensionLogs := observer.New(zapcore.ErrorLevel)
	collectorLogs := &observer.ObservedLogs{}
	collector := NewCollector(zap.New(extensionObservedCore), testFactories(t), "test")
	collector.coreFunc = func(levelEnabler zapcore.LevelEnabler) zapcore.Core {
		var collectorObservedCore zapcore.Core
		collectorObservedCore, collectorLogs = observer.New(levelEnabler)
		return collectorObservedCore
	}

	ctx := context.Background()
	err := collector.Start(ctx)
	require.NoError(t, err)

	err = collector.Stop()
	require.NoError(t, err)

	assert.NotEmpty(t, collectorLogs.FilterLevelExact(zapcore.InfoLevel).All(),
		"INFO logs from the collector should be emitted when collector config sets level: info")
	assert.Empty(t, extensionLogs.All(), "collector logs should not be written through the extension logger core")
}

func TestCollectorLogLevelDoesNotSuppressExtensionLogs(t *testing.T) {
	t.Setenv("OPENTELEMETRY_COLLECTOR_CONFIG_URI", "file:testdata/config-error-level.yaml")

	extensionObservedCore, extensionLogs := observer.New(zapcore.InfoLevel)
	collector := NewCollector(zap.New(extensionObservedCore), testFactories(t), "test")
	collector.coreFunc = func(levelEnabler zapcore.LevelEnabler) zapcore.Core {
		collectorObservedCore, _ := observer.New(levelEnabler)
		return collectorObservedCore
	}

	ctx := context.Background()
	err := collector.Start(ctx)
	require.NoError(t, err)

	collector.logger.Info("extension log")

	err = collector.Stop()
	require.NoError(t, err)

	assert.Len(t, extensionLogs.FilterMessage("extension log").All(), 1,
		"extension logs should be controlled by the extension logger, not collector config")
}

func TestCollectorConfigContentStartsCollector(t *testing.T) {
	yamlContent := `
receivers:
  nop:
exporters:
  nop:
service:
  telemetry:
    logs:
      level: info
  pipelines:
    traces:
      receivers: [nop]
      exporters: [nop]
`
	t.Setenv(envCollectorConfigContent, base64.StdEncoding.EncodeToString([]byte(yamlContent)))

	core, logs := observer.New(zapcore.InfoLevel)
	col := NewCollector(zap.New(core), testFactories(t), "test")

	ctx := context.Background()
	err := col.Start(ctx)
	require.NoError(t, err)

	err = col.Stop()
	require.NoError(t, err)

	assert.NotEmpty(t, logs.FilterMessage("Using inline config from "+envCollectorConfigContent).All())
}

func TestCollectorConfigURITakesPrecedenceOverContent(t *testing.T) {
	t.Setenv(envCollectorConfigURI, "file:testdata/config-info-level.yaml")
	t.Setenv(envCollectorConfigContent, base64.StdEncoding.EncodeToString([]byte("irrelevant: true")))

	core, logs := observer.New(zapcore.WarnLevel)
	col := NewCollector(zap.New(core), testFactories(t), "test")

	ctx := context.Background()
	err := col.Start(ctx)
	require.NoError(t, err)

	err = col.Stop()
	require.NoError(t, err)

	warnLogs := logs.FilterMessageSnippet(envCollectorConfigContent + " are set").All()
	assert.NotEmpty(t, warnLogs, "expected a warning about both env vars being set")
}

func TestCollectorConfigContentInvalidBase64FallsBack(t *testing.T) {
	t.Setenv(envCollectorConfigContent, "not-valid-base64!@#$")

	core, logs := observer.New(zapcore.ErrorLevel)
	col := NewCollector(zap.New(core), testFactories(t), "test")

	ctx := context.Background()
	// Collector will fall back to the default path which won't exist — expect start error
	err := col.Start(ctx)
	assert.Error(t, err, "expected collector to fail when falling back to missing default config")

	errorLogs := logs.FilterMessageSnippet("Failed to decode").All()
	assert.NotEmpty(t, errorLogs, "expected an error log about invalid Base64")
}

func testFactories(t *testing.T) otelcol.Factories {
	receivers, err := otelcol.MakeFactoryMap(receivertest.NewNopFactory())
	require.NoError(t, err)
	exporters, err := otelcol.MakeFactoryMap(exportertest.NewNopFactory())
	require.NoError(t, err)

	return otelcol.Factories{
		Receivers: receivers,
		Exporters: exporters,
		Telemetry: otelconftelemetry.NewFactory(),
	}
}
