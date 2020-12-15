# OpenTelemetry AWS Lambda Extension
*NOTE: This is an experimental AWS Lambda Extension for OpenTelemetry*

The OpenTelemetry Lambda Extension provides a mechanism to export telemetry aynchronously from AWS Lambdas. It does this by embedding an [OpenTelemetry Collector](https://github.com/open-telemetry/opentelemetry-collector) inside an [AWS Extension Layer](https://aws.amazon.com/blogs/compute/introducing-aws-lambda-extensions-in-preview/). This allows lambdas to use the OpenTelemetry Collector Exporter to send traces and metrics to any configured backend.

## Installing

To install the OpenTelemetry Lambda Extension to an existing Lambda function using the `aws` CLI:

```
aws lambda update-function-configuration --function-name Function --layers arn:aws:lambda:<AWS REGION>:297975325230:layer:opentelemetry-lambda-extension:8
```

Alternatively, to configure the OpenTelemetry Lambda Extension via SAM, add the following configuration:

```yaml
  Function:
    Type: AWS::Serverless::Function
    Properties:
      Layers:
        - arn:aws:lambda:<AWS REGION>:297975325230:layer:opentelemetry-lambda-extension:8
      ...
      Environment:
        Variables:
          OPENTELEMETRY_COLLECTOR_CONFIG_FILE: /var/task/collector.yaml
```

## Configuration

The OpenTelemetry Collector uses yaml for configuration. To configure the collector, add a `collector.yaml` to your function and specifiy its location via the `OPENTELEMETRY_COLLECTOR_CONFIG_FILE` environment file.

Here is a sample configuration file:

```yaml
receivers:
  otlp:
    protocols:
      grpc:
      http:
exporters:
  otlp:
    endpoint: destination:1234
    headers: {"header1":"value1"}
processors:
  batch:

service:
  extensions:
  pipelines:
    traces:
      receivers: [otlp]
      processors: [batch]
      exporters: [otlp]
```

Once the file has been deployed with a Lambda, configuring the `OPENTELEMETRY_COLLECTOR_CONFIG_FILE` will tell the OpenTelemetry extension where to find the collector configuration:

```
aws lambda update-function-configuration --function-name Function --environment Variables={OPENTELEMETRY_COLLECTOR_CONFIG_FILE=/var/task/collector.yaml}
```

You can configure environment variables via yaml as well:

```yaml
  Function:
    Type: AWS::Serverless::Function
    Properties:
      ...
      Environment:
        Variables:
          OPENTELEMETRY_COLLECTOR_CONFIG_FILE: /var/task/collector.yaml
```

