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

package logging

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

func TestNewLoggerWarnsForInvalidExtensionLogLevel(t *testing.T) {
	observedCore, logs := observer.New(zapcore.DebugLevel)

	logger := newLogger("not-a-level", func(zapcore.LevelEnabler) zapcore.Core {
		return observedCore
	})

	logger.Info("extension info")

	assert.Len(t, logs.FilterMessage("unable to parse log level from environment").All(), 1)
	assert.Len(t, logs.FilterMessage("extension info").All(), 1)
}

func TestNewLoggerAppliesValidExtensionLogLevel(t *testing.T) {
	logs := &observer.ObservedLogs{}
	logger := newLogger("error", func(levelEnabler zapcore.LevelEnabler) zapcore.Core {
		observedCore, observedLogs := observer.New(levelEnabler)
		logs = observedLogs
		return observedCore
	})

	logger.Info("extension info")
	logger.Error("extension error")

	assert.Empty(t, logs.FilterMessage("extension info").All())
	assert.Len(t, logs.FilterMessage("extension error").All(), 1)
}
