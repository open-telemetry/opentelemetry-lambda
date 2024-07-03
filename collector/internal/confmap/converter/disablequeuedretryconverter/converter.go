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
package disablequeuedretryconverter // import "github.com/open-telemetry/opentelemetry-lambda/collector/internal/confmap/converter/disablequeuedretryconverter"

import (
	"context"
	"fmt"
	"strings"

	"go.opentelemetry.io/collector/confmap"
)

const (
	expKey = "exporters"
)

var exporters = map[string]struct{}{
	"awskinesis":              {},
	"coralogix":               {},
	"datadog":                 {},
	"dynatrace":               {},
	"googlecloud":             {},
	"googlecloudpubsub":       {},
	"googlemanagedprometheus": {},
	"humio":                   {},
	"influxdb":                {},
	"jaeger":                  {},
	"kafka":                   {},
	"logzio":                  {},
	"loki":                    {},
	"mezmo":                   {},
	"observiq":                {},
	"opencensus":              {},
	"otlp":                    {},
	"otlphttp":                {},
	"pulsar":                  {},
	"sapm":                    {},
	"signalfx":                {},
	"skywalking":              {},
	"splunkhec":               {},
	"sumologic":               {},
	"tanzuobservability":      {},
	"zipkin":                  {},
}

type converter struct {
}

// New returns a confmap.Converter, that ensures queued retry is disabled for all configured exporters.
func New() confmap.Converter {
	return &converter{}
}

func (c converter) Convert(_ context.Context, conf *confmap.Conf) error {
	out := make(map[string]interface{})
	expVal := conf.Get(expKey)

	switch exps := expVal.(type) {
	case map[string]interface{}:
		for name := range exps {
			if _, ok := exporters[strings.Split(name, "/")[0]]; !ok {
				continue
			}
			out[fmt.Sprintf("%s::%s::sending_queue::enabled", expKey, name)] = false
		}
	}
	if err := conf.Merge(confmap.NewFromStringMap(out)); err != nil {
		return err
	}
	return nil
}
