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

package telemetryapireceiver

// EventType represents the type of log events in Lambda
type EventType string

const (
	// Platform is used to receive log events emitted by the Lambda platform
	Platform EventType = "platform"
	// PlatformInitStart is used when function initialization started.
	PlatformInitStart EventType = Platform + ".initStart"
	// PlatformInitRuntimeDone is used when function initialization ended.
	PlatformInitRuntimeDone EventType = Platform + ".initRuntimeDone"
	// Function is used to receive log events emitted by the function
	Function EventType = "function"
	// Extension is used is to receive log events emitted by the extension
	Extension EventType = "extension"
)

// BufferingCfg holds configuration for receiving telemetry from the Telemetry API.
// Telemetry will be sent to your listener when one of the conditions below is met.
type BufferingCfg struct {
	// Maximum number of log events to be buffered in memory. (default: 10000, minimum: 1000, maximum: 10000)
	MaxItems uint32 `json:"maxItems"`
	// Maximum size in bytes of the log events to be buffered in memory. (default: 262144, minimum: 262144, maximum: 1048576)
	MaxBytes uint32 `json:"maxBytes"`
	// Maximum time (in milliseconds) for a batch to be buffered. (default: 1000, minimum: 100, maximum: 30000)
	TimeoutMS uint32 `json:"timeoutMs"`
}

// URI is used to set the endpoint where the logs will be sent to
type URI string

// HTTPMethod represents the HTTP method used to receive events from the Telemetry API
type HTTPMethod string

const (
	// Receive log events via POST requests to the listener
	HttpPost HTTPMethod = "POST"
	// Receive log events via PUT requests to the listener
	HttpPut HTTPMethod = "PUT"
)

// Used to specify the protocol when subscribing to Telemetry API for HTTP
type HTTPProtocol string

const (
	HttpProto HTTPProtocol = "HTTP"
)

// Denotes what the content is encoded in
type HTTPEncoding string

const (
	JSON HTTPEncoding = "JSON"
)

// Configuration for listeners that would like to receive telemetry via HTTP
type Destination struct {
	Protocol   HTTPProtocol `json:"protocol"`
	URI        URI          `json:"URI"`
	HttpMethod HTTPMethod   `json:"method"`
	Encoding   HTTPEncoding `json:"encoding"`
}

type SchemaVersion string

// Request body that is sent to the Telemetry API on subscribe
type SubscribeRequest struct {
	SchemaVersion SchemaVersion `json:"schemaVersion"`
	EventTypes    []EventType   `json:"types"`
	BufferingCfg  BufferingCfg  `json:"buffering"`
	Destination   Destination   `json:"destination"`
}

type Event struct {
	Time   string `json:"time"`
	Type   string `json:"type"`
	Record any    `json:"record"`
}
