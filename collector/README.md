# OpenTelemetry Collector AWS Lambda Extension layer

The OpenTelemetry Collector Lambda Extension provides a mechanism to export telemetry aynchronously from AWS Lambdas. It does this by embedding a stripped-down version of [OpenTelemetry Collector Contrib](https://github.com/open-telemetry/opentelemetry-collector-contrib) inside an [AWS Extension Layer](https://aws.amazon.com/blogs/compute/introducing-aws-lambda-extensions-in-preview/). This allows lambdas to use the OpenTelemetry Collector Exporter to send traces and metrics to any configured backend.


## Build your OpenTelemetry Collector Lambda layer from scratch
At the moment users have to build Collector Lambda layer by themselves, we will provide sharing Lambda layer in the future.
- Download a local copy of the [opentelemetry-lambda repository from Github](https://github.com/open-telemetry/opentelemetry-lambda).
- Run command: `cd collector && make publish-layer` to publish OpenTelemetry Collector Lambda layer in your AWS account and get its ARN

Be sure to:

* Install [AWS CLI](https://docs.aws.amazon.com/cli/latest/userguide/install-cliv2.html)
* Config [AWS credential](https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-files.html)

## Installing
To install the OpenTelemetry Collector Lambda layer to an existing Lambda function using the `aws` CLI:

```
aws lambda update-function-configuration --function-name Function --layers <your Lambda layer ARN>
```

Alternatively, to configure the OpenTelemetry Lambda Extension via CloudFormation template, add the following configuration:

```yaml
  Function:
    Type: AWS::Serverless::Function
    Properties:
      Layers:
        - <your Lambda layer ARN>
      ...
```

## Configuration

By default, OpenTelemetry Collector Lambda layer exports telemetry data to AWS backends. To customize the collector configuration, add a `collector.yaml` to your function and specify its location via the `OPENTELEMETRY_COLLECTOR_CONFIG_FILE` environment file.

Here is a sample configuration file:

```yaml
receivers:
  otlp:
    protocols:
      grpc:

exporters:
  logging:
    loglevel: debug
  otlp:
    endpoint: { backend endpoint }

service:
  pipelines:
    traces:
      receivers: [otlp]
      exporters: [logging, otlp]
```

Once the file has been deployed with a Lambda, configuring the `OPENTELEMETRY_COLLECTOR_CONFIG_FILE` will tell the OpenTelemetry extension where to find the collector configuration:

```
aws lambda update-function-configuration --function-name Function --environment Variables={OPENTELEMETRY_COLLECTOR_CONFIG_FILE=/var/task/collector.yaml}
```

You can configure environment variables via CloudFormation template as well:

```yaml
  Function:
    Type: AWS::Serverless::Function
    Properties:
      ...
      Environment:
        Variables:
          OPENTELEMETRY_COLLECTOR_CONFIG_FILE: /var/task/collector.yaml
```

In addition to local files, the OpenTelemetry Collector Lambda layer may be configured through HTTP or S3 URIs
provided in the `OPENTELEMETRY_COLLECTOR_CONFIG_FILE` environment variable.  For instance, to load configuration
from an S3 object using a CloudFormation template:

```yaml
  Function:
    Type: AWS::Serverless::Function
    Properties:
      ...
      Environment:
        Variables:
          OPENTELEMETRY_COLLECTOR_CONFIG_FILE: s3://<bucket_name>.s3.<region>.amazonaws.com/collector_config.yaml
```

Loading configuration from S3 will require that the IAM role attached to your function includes read access to the relevant bucket.

## Auto-Configuration

Configuring the Lambda Collector without the decouple processor and batch processor can lead to performance issues. So the OpenTelemetry Lambda Layer automatically adds the decouple processor to the end of the chain if the batch processor is used and the decouple processor is not.

# Improving Lambda responses times
At the end of a lambda function's execution, the OpenTelemetry client libraries will flush any pending spans/metrics/logs
to the collector before returning control to the Lambda environment. The collector's pipelines are synchronous and this
means that the response of the lambda function is delayed until the data has been exported.
This delay can potentially be for hundreds of milliseconds.

To overcome this problem the [decouple](./processor/decoupleprocessor/README.md) processor can be used to separate the
two ends of the collectors pipeline and allow the lambda function to complete while ensuring that any data is exported
before the Lambda environment is frozen.

See the section regarding auto-configuration above. You don't need to manually add the decouple processor to your configuration.

## Reducing Lambda runtime
If your lambda function is invoked frequently it is also possible to pair the decouple processor with the batch
processor to reduce total lambda execution time at the expense of delaying the export of OpenTelemetry data.
When used with the batch processor the decouple processor must be the last processor in the pipeline to ensure that data
is successfully exported before the lambda environment is frozen.

As stated previously in the auto-configuration section, the OpenTelemetry Lambda Layer will automatically add the decouple processor to the end of the processors if the batch is used and the decouple processor is not. The result will be the same whether you configure it manually or not.
