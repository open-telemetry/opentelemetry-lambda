# Telemetry API Receiver

| Status                   |                  |
| ------------------------ | ---------------- |
| Stability                | [in development] |
| Supported pipeline types | traces, metrics  |
| Distributions            | [extension]      |

This receiver generates telemetry in response to events from the
[Telemetry API](https://docs.aws.amazon.com/lambda/latest/dg/telemetry-api.html). It does this by setting up an
endpoint and registering itself with the Telemetry API on startup. Generated telemetry includes both traces and
metrics.

## Traces

If used in a `traces` pipeline, this receiver currently creates two kinds of spans:

- A "FaaS invocation" span following the
  [semantic conventions for FaaS spans](https://opentelemetry.io/docs/specs/semconv/faas/faas-spans/). Unless the
  function is being initialized for the invocation, the span runs from the `platform.start` event to the
  `platform.runtimeDone` event. If the function is being initialized, the span's start timestamp is instead taken from
  the `platform.initStart` event.
- If the function is being initialized, another span named `faas.runtimeInit` is created as a child of the "FaaS
  invocation" span. This span runs from the `platform.initStart` to the `platform.initRuntimeDone` event.

In order to associate these spans with traces created during function invocation, consider using the
[`faasprocessor`](../../processor/faasprocessor/).

## Metrics

If used in a `metrics` pipeline, this receiver currently generates a all metrics specified in the
[semantic conventions for FaaS metrics](https://opentelemetry.io/docs/specs/semconv/faas/faas-metrics/) (except for
`faas.cpu_usage` which is not provided by the Telemetry API). Timestamps of all metrics are set to the timestamps
provided by certain events emitted by the Telemetry API such that metrics ought to capture timing information well.

Created metrics initialize counters to zero to prevent time series (such as `faas.coldstarts`) starting with a value
of 1. Otherwise, this would create issues for many systems (e.g. Prometheus) as the transition from value 0 to 1 would
never be observed.

## Configuration

The receiver can be enabled via the following configuration:

```yaml
receivers:
  telemetryapi:
    metrics:
      # `use_exponential_histograms` allows to generate histogram metrics using exponential buckets rather than
      # predefined ones. For Prometheus, exponential histograms currently requires enabling experimental features.
      use_exponential_histograms: false
```

[in development]: https://github.com/open-telemetry/opentelemetry-collector#development
[extension]: https://github.com/open-telemetry/opentelemetry-lambda/collector
