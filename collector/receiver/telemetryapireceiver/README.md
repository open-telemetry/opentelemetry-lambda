# Telemetry API Receiver

| Status                   |              |
| ------------------------ |--------------|
| Stability                | [alpha]      |
| Supported pipeline types | traces, logs |
| Distributions            | [extension]  |

This receiver generates telemetry in response to events from the [Telemetry API](https://docs.aws.amazon.com/lambda/latest/dg/telemetry-api.html). It does this by setting up an endpoint and registering itself with the Telemetry API on startup.

Supported events:

* `platform.initStart` - The receiver uses this event to record the start time of the function initialization period. Once both start and end times are recorded, the receiver generates a span named `platform.initRuntimeDone` to record the event.
* `platform.initRuntimeDone` - The receiver uses this event to record the end time of the function initialization period. Once both start and end times are recorded, the receiver generates a span named `platform.initRuntimeDone` to record the event.

## Configuration

| Field   | Default                               | Description                                                                                                                             |
|---------|---------------------------------------|-----------------------------------------------------------------------------------------------------------------------------------------|
| `port`  | 4235                                  | HTTP server port to receive Telemetry API data.                                                                                         |
| `types` | ["platform", "function", "extension"] | [Types](https://docs.aws.amazon.com/lambda/latest/dg/telemetry-api-reference.html#telemetry-subscribe-api) of telemetry to subscribe to |


```yaml
receivers:
    telemetryapi:
    telemetryapi/1:
      port: 4326
    telemetryapi/2:
      port: 4326
      types:
        - platform
        - function
    telemetryapi/3:
      port: 4326
      types: ["platform", "function"]
```

[alpha]: https://github.com/open-telemetry/opentelemetry-collector#alpha
[extension]: https://github.com/open-telemetry/opentelemetry-lambda/collector
