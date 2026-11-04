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

var (
	encoderConfig = zap.NewProductionEncoderConfig()
	stdoutSyncer  = zapcore.Lock(zapcore.AddSync(os.Stdout))
)

func NewLogger() *zap.Logger {
	lvl, err := parseLevel(os.Getenv(extensionLogLevelEnvVar))

	l := zap.New(NewCore(lvl))

	if err != nil {
		l.Warn("unable to parse log level from environment, falling back to default log level", zap.Error(err), zap.Stringer("default_level", lvl))
	}
	return l
}

func NewCore(levelEnabler zapcore.LevelEnabler) zapcore.Core {
	return zapcore.NewCore(zapcore.NewJSONEncoder(encoderConfig), stdoutSyncer, levelEnabler)
}

// parseLevel resolves the extension log level from the env var value,
// falling back to INFO and returning an error if the value is invalid.
func parseLevel(envLvl string) (zap.AtomicLevel, error) {
	if envLvl == "" {
		return zap.NewAtomicLevelAt(zapcore.InfoLevel), nil
	}
	userLvl, err := zap.ParseAtomicLevel(envLvl)
	if err != nil {
		return zap.NewAtomicLevelAt(zapcore.InfoLevel), err
	}
	return userLvl, nil
}
