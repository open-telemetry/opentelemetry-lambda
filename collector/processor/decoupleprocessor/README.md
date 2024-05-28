# Decouple Processor

| Status                   |                       |
| ------------------------ |-----------------------|
| Stability                | [alpha]               |
| Supported pipeline types | traces, metrics, logs |
| Distributions            | [extension]           |

This processor decouples the receiver and exporter ends of the pipeline. This allows the lambda function to finish before traces/metrics/logs have been exported by the collector. The processor is aware of the Lambda [lifecycle] and will prevent the environment from being frozen or shutdown until any pending traces/metrics/logs have been exported.
In this way the response times of the Lambda function is not impacted by the need to export data, however the billed duration will include the time taken to export data as well as runtime of the lambda function.

The decouple processor should always be the last processor in the list to ensure that there are no issues with data being sent while the environment is about to be frozen, which could result in lost data.

When combined with the batch processor, the number of exports required can be significantly reduced and therefore the cost of running the lambda. This is with the trade-off that the data will not be available at your chosen endpoint until some time after the invocation, up to a maximum of 5 minutes (the timeout that the environment is shutdown when no further invocations are received).

## Auto-Configuration

Due to the significant performance improvements with this approach, the OpenTelemetry Lambda Layer automatically configures the decouple processor when the batch processor is used. This ensures the best performance by default.

When running the Collector for the Lambda Layer, the configuration is converted by automatically adding the decouple processor to all pipelines if the following conditions are met:

1. The pipeline contains a batch processor.
2. There is no decouple processor already defined after the batch processor.

This automatic configuration helps prevent the data loss scenarios that can occur when the Lambda environment is frozen as the batch processor continues aggregating data. The decouple processor allows the Lambda function invocation to complete while the collector continues exporting the data asynchronously.

## Processor Configuration

```yaml
processors:
    decouple:
      # max_queue_size allows you to control how many spans etc. are accepted before the pipeline blocks
      # until an export has been completed. Default value is 200.
      max_queue_size:  20
```

[alpha]: https://github.com/open-telemetry/opentelemetry-collector#development
[extension]: https://github.com/open-telemetry/opentelemetry-lambda/collector
[lifecycle]: https://docs.aws.amazon.com/lambda/latest/dg/runtimes-extensions-api.html#runtimes-extensions-api-lifecycle
