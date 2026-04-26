# Telemetry API Receiver

| Status                   |                       |
|--------------------------|-----------------------|
| Stability                | [alpha]               |
| Supported pipeline types | traces, logs, metrics |
| Distributions            | [extension]           |

This receiver generates telemetry in response to events from the [Telemetry API](https://docs.aws.amazon.com/lambda/latest/dg/telemetry-api.html). It does this by setting up an endpoint and registering itself with the Telemetry API on startup.

Supported events:

* `platform.initStart` - The receiver uses this event to record the start time of the function initialization period. Once both start and end times are recorded, the receiver generates a span named `platform.initRuntimeDone` to record the event.
* `platform.initRuntimeDone` - The receiver uses this event to record the end time of the function initialization period. Once both start and end times are recorded, the receiver generates a span named `platform.initRuntimeDone` to record the event.

## Logs metadata reserved fields

The following field names are reserved for internal use in logs metadata and must not be used as custom metadata keys:

| Field       | Description                              |
|-------------|------------------------------------------|
| `level`     | Log severity level                       |
| `message`   | Log message body                         |
| `requestId` | AWS Lambda request identifier            |
| `timestamp` | Time of the log event                    |
| `type`      | Telemetry API event type                 |

> **Note:** These fields are populated internally by the receiver and will take priority over any user-provided metadata with the same name. You should be aware of these limitations and handle conflicts at the business logic level — for example, by renaming custom fields that collide with reserved names before they are emitted as log metadata.

## Configuration

| Field                 | Default                               | Description                                                                                                                                                          |
|-----------------------|---------------------------------------|----------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `port`                | 0 (dynamically determined by OS)      | HTTP server port to receive Telemetry API data.                                                                                                                      |
| `types`               | ["platform", "function", "extension"] | [Types](https://docs.aws.amazon.com/lambda/latest/dg/telemetry-api-reference.html#telemetry-subscribe-api) of telemetry to subscribe to                              |
| `metrics_temporality` | cumulative                            | The [aggregation temporality](https://opentelemetry.io/docs/specs/otel/metrics/data-model/#temporality) to use for metrics. Supported values: `delta`, `cumulative`. |
| `export_interval_ms`  | 60000                                 | The interval in milliseconds at which metrics are exported. If set to 0, metrics are exported immediately upon receipt.                                              |


```yaml
receivers:
    telemetryapi:
    telemetryapi/1:
      port: 4326
      export_interval_ms: 30000
    telemetryapi/2:
      port: 4327
      types:
        - platform
        - function
      metrics_temporality: delta
    telemetryapi/3:
      port: 4328
      types: ["platform", "function"]
```

[alpha]: https://github.com/open-telemetry/opentelemetry-collector#alpha
[extension]: https://github.com/open-telemetry/opentelemetry-lambda/tree/main/collector
