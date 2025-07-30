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

package telemetryapireceiver

import (
	"log"
	"time"
)

type event struct {
	Time   string `json:"time"`
	Type   string `json:"type"`
	Record any    `json:"record"`
}

// getTime parses the event's timestamp string into a time.Time object.
func (e *event) getTime() time.Time {
	t, err := time.Parse(time.RFC3339, e.Time)
	if err != nil {
		log.Printf("WARN: Failed to parse event timestamp '%s', using current time as fallback: %v", e.Time, err)
		return time.Now()
	}
	return t
}
