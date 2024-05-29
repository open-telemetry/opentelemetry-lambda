# Telemetry API Receiver

| Status                   |                       |
| ------------------------ |-----------------------|
| Stability                | [in development]      |
| Supported pipeline types | traces, metrics, logs |
| Distributions            | [extension]           |

This receiver generates telemetry in response to events from the [Telemetry API](https://docs.aws.amazon.com/lambda/latest/dg/telemetry-api.html). It does this by setting up an endpoint and registering itself with the Telemetry API on startup.

Supported events:

* `platform.initStart` - The receiver uses this event to record the start time of the function initialization period. Once both start and end times are recorded, the receiver generates a span named `platform.initRuntimeDone` to record the event.
* `platform.initRuntimeDone` - The receiver uses this event to record the end time of the function initialization period. Once both start and end times are recorded, the receiver generates a span named `platform.initRuntimeDone` to record the event.

## Configuration

There are currently no configuration parameters available for this receiver. It can be enabled via the following configuration:

```yaml
receivers:
    telemetryapi:
```

[in development]: https://github.com/open-telemetry/opentelemetry-collector#development
[extension]: https://github.com/open-telemetry/opentelemetry-lambda/collector
