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

type event struct {
	Time   string         `json:"time"`
	Type   string         `json:"type"`
	Record map[string]any `json:"record"`
}

// NOTE: Types defined here do not include all attributes sent by the Telemetry API but only those
//       relevant to this package. For a full overview, consult the documentation:
//       https://docs.aws.amazon.com/lambda/latest/dg/telemetry-schema-reference.html#telemetry-api-events

type platformInitReportRecord struct {
	Metrics struct {
		DurationMs float64 `mapstructure:"durationMs"`
	} `mapstructure:"metrics"`
}

type platformStartRecord struct {
	RequestID string `mapstructure:"requestId"`
}

type platformRuntimeDoneRecord struct {
	RequestID string `mapstructure:"requestId"`
	Status    status `mapstructure:"status"`
	Metrics   *struct {
		DurationMs float64 `mapstructure:"durationMs"`
	} `mapstructure:"metrics"`
}

type platformReport struct {
	Metrics struct {
		MaxMemoryUsedMb int64 `mapstructure:"maxMemoryUsedMB"`
	} `mapstructure:"metrics"`
}

/* ----------------------------------------- CONSTANTS ----------------------------------------- */

type status string

const (
	statusSuccess = status("success")
	statusFailure = status("failure")
	statusError   = status("error")
	statusTimeout = status("timeout")
)
