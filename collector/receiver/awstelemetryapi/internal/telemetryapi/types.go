package telemetryapi

// EventType represents the type of log events in Lambda
type EventType string

// Constants for all Telemetry API event types we handle.
const (
	Platform                EventType = "platform"
	PlatformInitStart       EventType = "platform.initStart"
	PlatformInitRuntimeDone EventType = "platform.initRuntimeDone"
	PlatformStart           EventType = "platform.start"
	PlatformRuntimeDone     EventType = "platform.runtimeDone"
	PlatformReport          EventType = "platform.report"
	Function                EventType = "function"
	Extension               EventType = "extension"
)

// BufferingCfg holds configuration for the subscription buffer.
type BufferingCfg struct {
	MaxItems  uint `json:"maxItems"`
	MaxBytes  uint `json:"maxBytes"`
	TimeoutMS uint `json:"timeoutMs"`
}

// Destination is where the Telemetry API will send telemetry.
type Destination struct {
	Protocol string `json:"protocol"`
	URI      string `json:"URI"`
}

// RegisterRequest is the request body for the /extension/register endpoint.
type RegisterRequest struct {
	Events []string `json:"events"`
}

// SubscribeRequest is the request body for the /telemetry endpoint.
type SubscribeRequest struct {
	SchemaVersion string       `json:"schemaVersion"`
	Types         []EventType  `json:"types"`
	Buffering     BufferingCfg `json:"buffering"`
	Destination   Destination  `json:"destination"`
}
