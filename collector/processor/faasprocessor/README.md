# FaaS Processor

| Status                   |                  |
| ------------------------ | ---------------- |
| Stability                | [in development] |
| Supported pipeline types | traces           |
| Distributions            | [extension]      |

This processor associates spans created by the [telemetryapireceiver](../../receiver/telemetryapireceiver) with
incoming span data processed by the Collector extension. To this end, it searches for a pair of spans with the same
value for the `faas.invocation_id` attribute:

- The first span is created by the `telemetryapireceiver` and can easily be identified via its scope.
- The second span must be created by the user application and be received via the collector extension.

Once a matching pair is found, the span created by the `telemetryapireceiver` is updated to belong to the same trace as
the user-created span and is "inserted" into the span hierarchy: the span created by the `telemetryapireceiver` is set
as the parent span of the user-created span and is itself set as child of the previous user-created span's parent (if
any). If the `telemetryapireceiver` also created a span for the function initialization, it is simply updated to belong
to the same trace and remains a child of the created "FaaS invocation span".

**Note:** If your application does not emit any spans with the `faas.invocation_id` attribute set, DO NOT use this
processor. It will store all traces trying to search for matches of this attribute and never emits them (until the
Lambda runtime shuts down in which case all unmatched spans are emitted).

There are currently no configuration parameters available for this processor. It can be enabled via the following
configuration:

```yaml
processors:
  faas:
```

[in development]: https://github.com/open-telemetry/opentelemetry-collector#development
[extension]: https://github.com/open-telemetry/opentelemetry-lambda/collector
