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

// Package faasprocessor associates spans created by the telemetryapireceiver with incoming span
// data processed by the Collector extension. To this end, it searches for a pair of spans with
// the same value for the faas.invocation_id attribute:
//
//   - The first span is created by the telemetryapireceiver and can easily be identified via its
//     scope.
//   - The second span must be created by the user application and be received via the collector
//     extension.
//
// Once a matching pair is found, the span created by the telemetryapireceiver is updated to
// belong to the same trace as the user-created span and is "inserted" into the span hierarchy:
// the span created by the telemetryapireceiver is set as the parent span of the user-created span
// and is itself set as child of the previous user-created span's parent (if any). If the
// telemetryapireceiver also created a span for the function initialization, it is simply updated
// to belong to the same trace and remains a child of the created "FaaS invocation span".
//
// NOTE: If your application does not emit any spans with the faas.invocation_id attribute set, DO
// NOT use this processor. It will store all traces trying to search for matches of this attribute
// and never emits them (until the Lambda runtime shuts down in which case all unmatched spans are
// emitted).
package faasprocessor // import "github.com/open-telemetry/opentelemetry-lambda/collector/processor/faasprocessor"
