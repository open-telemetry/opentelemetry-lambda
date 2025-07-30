# AWS Telemetry API Receiver

| Status                   |                         |
| ------------------------ |-------------------------|
| Stability                | [alpha][alpha]            |
| Supported pipeline types | traces, logs, metrics   |
| Distributions            | [extension][extension]  |

This receiver generates telemetry in response to events from the [AWS Lambda Telemetry API](https://docs.aws.amazon.com/lambda/latest/dg/telemetry-api.html). It runs as a Lambda Extension, setting up an endpoint and registering itself with the Telemetry API on startup to convert platform and function events into OpenTelemetry signals.

### Supported Events

The receiver processes events from the Telemetry API and converts them as follows:

  * **Metrics**: `platform.report` events are converted into OTel Gauge metrics. This includes key performance indicators like `durationMs`, `billedDurationMs`, `memorySizeMB`, and `maxMemoryUsedMB`.
  * **Traces**: Lifecycle events are used to create spans:
      * `platform.initStart` and `platform.initRuntimeDone` are used to create a span for the function initialization phase (cold start).
      * `platform.start` and `platform.runtimeDone` are used to create a span for the function invocation phase.
  * **Logs**: `function` and `extension` events are converted into OTel Log records, preserving the original message, timestamp, and severity.

## Configuration

The following settings can be configured:

| Field       | Default                                 | Description                                                                                                                                    |
|-------------|-----------------------------------------|------------------------------------------------------------------------------------------------------------------------------------------------|
| `port`      | `4325`                                  | The HTTP server port to receive Telemetry API data.                                                                                             |
| `types`     | `["platform", "function", "extension"]` | The [types](https://docs.aws.amazon.com/lambda/latest/dg/telemetry-api-reference.html#telemetry-subscribe-api) of telemetry to subscribe to. |
| `maxItems`  | `1000`                                  | The maximum number of events to buffer in Lambda's memory before sending.                                                                       |
| `maxBytes`  | `262144`                                | The maximum size (in bytes) of events to buffer in Lambda's memory before sending.                                                              |
| `timeoutMs` | `1000`                                  | The maximum time (in milliseconds) to buffer events before sending.                                                                            |

### Example Configuration

```yaml
receivers:
  telemetryapireceiver:
    port: 4325
    types:
      - platform
      - function
    maxItems: 2000
    maxBytes: 524288
    timeoutMs: 500
```

[alpha]: https://github.com/open-telemetry/opentelemetry-collector#alpha
[extension]: https://github.com/open-telemetry/opentelemetry-lambda/tree/main/collector