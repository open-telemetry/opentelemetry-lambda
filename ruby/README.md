# OpenTelemetry Lambda Ruby

Scripts and files used to build AWS Lambda Layers for running OpenTelemetry on AWS Lambda for Ruby.

**Requirement**
* [Ruby 3.2.0](https://www.ruby-lang.org/en/news/2022/12/25/ruby-3-2-0-released/) (only supported version)
* [SAM CLI](https://docs.aws.amazon.com/serverless-application-model/latest/developerguide/serverless-sam-cli-install.html)
* [AWS CLI](https://docs.aws.amazon.com/cli/latest/userguide/install-cliv2.html)
* [Go](https://go.dev/doc/install)
* [Docker](https://docs.docker.com/get-docker)

**Building Lambda Ruby Layer With OpenTelemetry Ruby Dependencies**

1. Pull and install all the gem dependencies in to `.aws-sam` folder

```bash
sam build -u -t template.yml
```

2. Zip all the gems file, wrapper and handler into single zip file

```bash
(cd .aws-sam/build/OTelLayer/ && zip -qr ../<your_layer_name>.zip .)
mv .aws-sam/build/<your_layer_name>.zip .

# Or run the script
zip_ruby_layer.sh -n <your_layer_name>
```

**Default GEM_PATH**

The [default GEM_PATH](https://docs.aws.amazon.com/lambda/latest/dg/ruby-package.html#ruby-package-dependencies-layers) for aws lambda ruby is `/opt/ruby/gems/<ruby_vesion>` after lambda function loads this layer.

**Define AWS_LAMBDA_EXEC_WRAPPER**

Point `AWS_LAMBDA_EXEC_WRAPPER` to `/opt/otel-handler` to take advantage of layer wrapper that load all opentelemetry ruby components
e.g.
```
AWS_LAMBDA_EXEC_WRAPPER: /opt/otel-handler
```

#### There are two ways to define the AWS_LAMBDA_EXEC_WRAPPER that point to either binary executable or script (normally bash).

Method 1: define the AWS_LAMBDA_EXEC_WRAPPER in function from template.yml
```yaml
AWSTemplateFormatVersion: '2010-09-09'
Transform: 'AWS::Serverless-2016-10-31'
Description: OpenTelemetry Ruby Lambda layer for Ruby
Parameters:
  LayerName:
    ...
Resources:
  OTelLayer:
    ...
  api:
    ...
  function:
    Type: AWS::Serverless::Function
    Properties:
      ...
      Environment:
        Variables:
          AWS_LAMBDA_EXEC_WRAPPER: /opt/otel-handler # this is an example of the path

```

Method 2: directly update the environmental variable in lambda console: Configuration -> Environemntal variables

For more information about aws lambda wrapper and wrapper layer, check [aws lambda runtime-wrapper](https://docs.aws.amazon.com/lambda/latest/dg/runtimes-modify.html#runtime-wrapper). We provide a sample wrapper file in `src/layer/otel-handler` as reference.

### Sample App

1. Make sure the requirements are met (e.g. sam, aws, docker, ruby version.)
2. Navigate to the path `cd ruby/sample-apps`
3. Build the layer and function based on template.yml. You will see .aws-sam folder after executed the command
```bash
sam build -u -t template.yml
# for different arch, define it in properties from template.yml
   # Architectures:
    #   - arm64
```
4. Test with local simulation
```bash
sam local start-api --skip-pull-image
```

5. curl the lambda function
```bash
curl http://127.0.0.1:3000
# you should expect: Hello 1.4.1
```
In this sample-apps, we use `src/layer/otel-handler` as default `AWS_LAMBDA_EXEC_WRAPPER`; to change it, please edit in `sample-apps/template.yml`

In `ruby/sample-apps/template.yml`, the OTelLayer -> Properties -> ContentUri is pointing to `ruby/src/layer/`. This is for local testing purpose. If you wish to deploy (e.g. `sam deploy`), please point it to correct location or zip file.

### Test with Jaeger Endpoint

Assume you have a lambda function with current [released layer](https://github.com/open-telemetry/opentelemetry-lambda/releases/tag/layer-ruby%2F0.1.0), and you want to test it out that send trace to jaeger endpoint, below should be your environmental variable.
```
AWS_LAMBDA_EXEC_WRAPPER=/opt/otel-handler
OTEL_EXPORTER_OTLP_TRACES_ENDPOINT=http://<jaeger_endpoint:port_number>/v1/traces
```
Try with `jaeger-all-in-one` at [Jaeger](https://www.jaegertracing.io/docs/1.57/getting-started/)


