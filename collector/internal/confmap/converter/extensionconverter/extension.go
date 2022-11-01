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

package extensionconverter // import "github.com/open-telemetry/opentelemetry-lambda/collector/internal/confmap/converter/extensionconverter"

import (
	"context"
	"go.opentelemetry.io/collector/confmap"
)

const (
	serviceKey = "service::extensions"
	extKey     = "extensions"
)

type converter struct {
	extensions map[string]interface{}
}

// New returns a confmap.Converter, that ensures the lambda extension is configured
func New(extensions map[string]interface{}) confmap.Converter {
	return &converter{
		extensions: extensions,
	}
}

func (c converter) Convert(_ context.Context, conf *confmap.Conf) error {
	for name, extConf := range c.extensions {
		out := make(map[string]interface{})
		svcVal := conf.Get(serviceKey)
		extVal := conf.Get(extKey)
		switch v := svcVal.(type) {
		case string:
			out[serviceKey] = []interface{}{v, name}
		case []interface{}:
			out[serviceKey] = append(v, name)
		default:
			out[serviceKey] = []interface{}{name}
		}

		switch v2 := extVal.(type) {
		case map[string]interface{}:
			v2[name] = extConf
			out[extKey] = v2
		default:
			out[extKey] = map[string]interface{}{name: extConf}
		}
		if err := conf.Merge(confmap.NewFromStringMap(out)); err != nil {
			return err
		}
	}
	return nil
}
