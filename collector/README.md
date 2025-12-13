# OpenTelemetry Collector AWS Lambda Extension layer

The OpenTelemetry Collector Lambda Extension provides a mechanism to export telemetry aynchronously from AWS Lambdas. It does this by embedding a stripped-down version of [OpenTelemetry Collector Contrib](https://github.com/open-telemetry/opentelemetry-collector-contrib) inside an [AWS Extension Layer](https://aws.amazon.com/blogs/compute/introducing-aws-lambda-extensions-in-preview/). This allows lambdas to use the OpenTelemetry Collector Exporter to send traces and metrics to any configured backend.


## Build your OpenTelemetry Collector Lambda layer from scratch
At the moment users have to build Collector Lambda layer by themselves, we will provide sharing Lambda layer in the future.
- Download a local copy of the [opentelemetry-lambda repository from Github](https://github.com/open-telemetry/opentelemetry-lambda).
- Run command: `cd collector && make publish-layer` to publish OpenTelemetry Collector Lambda layer in your AWS account and get its ARN

Be sure to:

* Install [AWS CLI](https://docs.aws.amazon.com/cli/latest/userguide/install-cliv2.html)
* Config [AWS credential](https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-files.html)

## (Experimental) Customized collector build
The collector can be built with a customized set of connectors/exporters/receivers/processors. This feature is **experimental** and is only supported for self-built binaries of the collector.

### Build Tags
The build-tag `lambdacomponents.custom` must always be provided to opt-in for a custom build.
Once this build-tag is present, you need provide additional build-tags to include your desired components in the resulting binary:

- `lambdacomponents.all` includes all available components
- `lambdacomponents.connector.all` includes all available connectors
- `lambdacomponents.exporter.all` includes all available exporters
- `lambdacomponents.extension.all` includes all available extensions
- `lambdacomponents.processor.all` includes all available processors
- `lambdacomponents.receiver.all` includes all available receivers

Each available component can also be included explicitly by using its specific build-tag. For a full-list of available components, have a look into the [lambdacomponents](./lambdacomponents) package.

As an example, the full command to publish OpenTelemetry Collector Lambda layer in your AWS account and get its ARN including the following components:
- All receivers
- All processors
- No extensions
- Only the otlphttp exporter
- Only the spanmetrics connector

would be the following:
```shell
cd collector && BUILDTAGS="lambdacomponents.custom,lambdacomponents.receiver.all,lambdacomponents.processor.all,lambdacomponents.exporter.otlphttp,lambdacomponents.connector.spanmetrics" make publish-layer
```

### Adding additional options
To add more options for a customized build, you can add your desired component to the [lambdacomponents](./lambdacomponents) package.
Make sure to always restrict your addition using the appropriate build-tags.

For example, if you want to add the extension `foo`, the file providing this extension should be located in the [extension](./lambdacomponents/extension) directory have the following build restriction:
```
//go:build lambdacomponents.custom && (lambdacomponents.all || lambdacomponents.extension.all || lambdacomponents.extension.foo)
```

You can provide your addition as a pull-request to this repository. Before doing so, please also read through the details of [Contributing](#contributing) to this project.

## Build and publish your own OpenTelemetry Collector Lambda layer

To build and publish collector Lambda layer from your own fork into your own AWS account, 
you can use the `Publish Collector Lambda Layer` workflow which can only be triggered manually.


To do that, first you need to 
- Create Github's OIDC provider in your (or target) AWS account (for more details, you can check [here](https://github.com/aws-actions/configure-aws-credentials?oidc))
- Create an AWS IAM Role in the AWS account to be assumed by the `Publish Collector Lambda Layer` workflow from your forked OpenTelemetry Lambda repository.

To setup those, you can use (copy or load) the AWS CloudFormation template [here](../utils/aws-cloudformation/aws-cf-stack-for-layer-publish.yml).
Once AWS CloudFormation stack is created from the given template, 
ARN of the created AWS IAM Role to be assumed will be shown as `RoleARN` in the output of the stack, so note it to be used later.

After that, you can run the `Publish Collector Lambda Layer` workflow to build the Lambda collector and publish it to the target AWS account as Lambda layer: 
- Specify the architecture of the collector Lambda layer to be published via the `Architecture of the layer to be published` input. 
  Available options are `all`, `amd64` and `arm64`.
  The default value is `all` which builds and publishes layer for both of the `amd64` and `arm64` architectures.
- Specify the AWS region(s) where the collector Lambda layer will be published to via the `AWS Region(s) where layer will be published` input.
  Available options are `all`, `ap-northeast-1`, `ap-northeast-2`, `ap-south-1`, `ap-southeast-1`, `ap-southeast-2`, `ca-central-1`, `ca-west-1`, `eu-central-1`, `eu-north-1`, `eu-west-1`, `eu-west-2`, `eu-west-3`, `sa-east-1`, `us-east-1`, `us-east-2`, `us-west-1`, `us-west-2`.
  The default value is `all` which publishes layer to all the defined AWS regions mentioned above.
- Specify the AWS IAM Role ARN to be assumed for publishing layer via the `AWS IAM Role ARN to be assumed for publishing layer` input.
  This is the ARN of the AWS IAM Role you have taken from the `RoleARN` output variable of the created AWS CloudFormation stack above.
  This input is **optional** and if not specified, AWS IAM Role ARN to be assumed is tried to be resolved from `OTEL_LAMBDA_LAYER_PUBLISH_ROLE_ARN` secret.
  If it is still not able to resolved (neither this input is specified, nor `OTEL_LAMBDA_LAYER_PUBLISH_ROLE_ARN` secret is defined), 
  layer publish job will fail due to missing AWS credentials.
- Specify the layer version to be appended into layer name via the `Layer version to be appended into the layer name` input 
  to be used in the following format: `opentelemetry-lambda-collector-${architecture}-${layer-version}`.
  This input is **optional** and if not specified, layer name is generated in the `opentelemetry-lambda-collector-${architecture}` format without layer version postfix.
- Specify the build tags to build the collector with a customized set of connectors/exporters/receivers/processors 
  via the `Build tags to customize collector build` input.
  This input is **optional** and if not specified, collector is built with the default set of connectors/exporters/receivers/processors.
  Check the [Build Tags](#build-tags) section for the details.

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

By default, OpenTelemetry Collector Lambda layer exports telemetry data to AWS backends. To customize the collector configuration, add a `collector.yaml` to your function and specify its location via the `OPENTELEMETRY_COLLECTOR_CONFIG_URI` environment file.

Here is a sample configuration file:

```yaml
receivers:
  otlp:
    protocols:
      grpc:

exporters:
  debug:
    verbosity: detailed
  otlp:
    endpoint: { backend endpoint }

service:
  pipelines:
    traces:
      receivers: [otlp]
      exporters: [debug, otlp]
```

Once the file has been deployed with a Lambda, configuring the `OPENTELEMETRY_COLLECTOR_CONFIG_URI` will tell the OpenTelemetry extension where to find the collector configuration:

```
aws lambda update-function-configuration --function-name Function --environment Variables={OPENTELEMETRY_COLLECTOR_CONFIG_URI=/var/task/collector.yaml}
```

You can configure environment variables via CloudFormation template as well:

```yaml
  Function:
    Type: AWS::Serverless::Function
    Properties:
      ...
      Environment:
        Variables:
          OPENTELEMETRY_COLLECTOR_CONFIG_URI: /var/task/collector.yaml
```

In addition to local files, the OpenTelemetry Collector Lambda layer may be configured through HTTP or S3 URIs
provided in the `OPENTELEMETRY_COLLECTOR_CONFIG_URI` environment variable.  For instance, to load configuration
from an S3 object using a CloudFormation template:

```yaml
  Function:
    Type: AWS::Serverless::Function
    Properties:
      ...
      Environment:
        Variables:
          OPENTELEMETRY_COLLECTOR_CONFIG_URI: s3://<bucket_name>.s3.<region>.amazonaws.com/collector_config.yaml
```

Loading configuration from S3 will require that the IAM role attached to your function includes read access to the relevant bucket.

## Environment Variables

The following environment variables can be used to configure the OpenTelemetry Collector Lambda extension:

| Variable Name                        | Value                                                                          | Description                                                                                                                                                                                                                                                 |
| ------------------------------------ | ------------------------------------------------------------------------------ | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `OPENTELEMETRY_COLLECTOR_CONFIG_URI` | URI (e.g., `/var/task/collector.yaml`, `http://...`, `s3://...`)               | Specifies the location of the OpenTelemetry Collector configuration file. This can be a path within the function's deployment package, an HTTP URI, or an S3 URI. If loading from S3, the function's IAM role needs read access to the specified S3 object. |
| `OPENTELEMETRY_EXTENSION_LOG_LEVEL`  | `debug`, `info`, `warn`, `error`, `dpanic`, `panic`, `fatal` (Default: `info`) | Controls the logging level of the OpenTelemetry Lambda extension itself.                                                                                                                                                                                    |

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
