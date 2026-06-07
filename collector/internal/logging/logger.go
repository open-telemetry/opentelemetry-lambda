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
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const extensionLogLevelEnvVar = "OPENTELEMETRY_EXTENSION_LOG_LEVEL"

type CoreFunc func(zapcore.LevelEnabler) zapcore.Core

var (
	encoderConfig = zap.NewProductionEncoderConfig()
	stdoutSyncer  = zapcore.Lock(zapcore.AddSync(os.Stdout))
)

func NewLogger() *zap.Logger {
	return newLogger(os.Getenv(extensionLogLevelEnvVar), NewCore)
}

func NewCore(levelEnabler zapcore.LevelEnabler) zapcore.Core {
	return zapcore.NewCore(zapcore.NewJSONEncoder(encoderConfig), stdoutSyncer, levelEnabler)
}

func newLogger(envLvl string, coreFunc CoreFunc) *zap.Logger {
	lvl := zap.NewAtomicLevelAt(zapcore.InfoLevel)
	var err error
	if envLvl != "" {
		var userLvl zap.AtomicLevel
		userLvl, err = zap.ParseAtomicLevel(envLvl)
		if err == nil {
			lvl = userLvl
		}
	}

	l := zap.New(coreFunc(lvl))

	if err != nil && envLvl != "" {
		l.Warn("unable to parse log level from environment", zap.Error(err))
	}

	return l
}
