// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package deprecatedprocessorconverter // import "github.com/open-telemetry/opentelemetry-lambda/collector/internal/confmap/converter/deprecatedprocessorconverter"

import (
	"context"
	"sort"
	"strings"

	"go.opentelemetry.io/collector/confmap"
	"go.uber.org/zap"
)

const procKey = "processors"

// deprecatedProcessors lists the processor types that are deprecated in the
// default collector layer build and will be removed in the next release.
var deprecatedProcessors = map[string]struct{}{
	"attributes": {},
	"resource":   {},
	"span":       {},
}

type converter struct {
	logger *zap.Logger
}

// New returns a confmap.Converter that logs a warning for every configured
// processor that is deprecated in the default collector layer build. It does
// not modify the configuration.
func New(logger *zap.Logger) confmap.Converter {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &converter{logger: logger}
}

func (c converter) Convert(_ context.Context, conf *confmap.Conf) error {
	procs, ok := conf.Get(procKey).(map[string]any)
	if !ok {
		return nil
	}

	var found []string
	for name := range procs {
		if _, ok := deprecatedProcessors[strings.Split(name, "/")[0]]; ok {
			found = append(found, name)
		}
	}
	sort.Strings(found)

	for _, name := range found {
		c.logger.Warn("Processor is deprecated in the default collector layer build and will be removed in the next release. "+
			"Migrate to the transform processor or build a custom layer that includes it. "+
			"See https://github.com/open-telemetry/opentelemetry-lambda/blob/layer-collector/0.23.0/collector/README.md#deprecation-notice",
			zap.String("processor", name))
	}
	return nil
}
