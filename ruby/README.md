# OpenTelemetry Lambda Ruby

Scripts and files used to build AWS Lambda Layers for running OpenTelemetry on AWS Lambda for Ruby.

Requirement:
* [Ruby 3.2.0](https://www.ruby-lang.org/en/news/2022/12/25/ruby-3-2-0-released/)
* [SAM CLI](https://docs.aws.amazon.com/serverless-application-model/latest/developerguide/serverless-sam-cli-install.html)
* [AWS CLI](https://docs.aws.amazon.com/cli/latest/userguide/install-cliv2.html)
* [Go](https://go.dev/doc/install)
* [Docker](https://docs.docker.com/get-docker)


## Building Lambda Ruby Layer With OpenTelemetry Ruby Dependencies

Build
```bash
sam build -u -t template.yml
```

Make sure the layer structure like below with your zip
```
ruby
    └── gems
        └── 3.2.0
            ├── build_info
            ├── doc
            ├── extensions
            ├── gems
            ├── plugins
            └── specifications
```

Zip the file for uploading to ruby lambda layer

```bash
cd .aws-sam/build/OTelLayer/
zip -qr ../../../<your_layer_name>.zip ruby/
cd -

# or run following script
zip_ruby_layer.sh -n <your_layer_name>
```

### Why build layer this way

The default GEM_PATH for Lambda Ruby is $LAMBDA_TASK_ROOT/vendor/bundle/ruby/2.5.0:/opt/ruby/gems/2.5.0. Therefore, it's important to set the Ruby path in the format `/opt/ruby/gems/<ruby_version>`.

Reference for build ruby lambda layer
https://docs.aws.amazon.com/lambda/latest/dg/ruby-package.html


### Define the AWS_LAMBDA_EXEC_WRAPPER

There are two ways to define the AWS_LAMBDA_EXEC_WRAPPER that point to either binary executable or script (normally bash).

#### Method 1: define the AWS_LAMBDA_EXEC_WRAPPER in function from template.yml
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

#### Method 2: directly update the environmental variable in lambda console: Configuration -> Environemntal variables 

For more information about aws lambda wrapper and wrapper layer, check [aws lambda runtime-wrapper](https://docs.aws.amazon.com/lambda/latest/dg/runtimes-modify.html#runtime-wrapper). We provide a sample wrapper file in `src/layer/otel-handler` as reference.

### Tracing and export trace to collector

To enable the default opentelemetry tracing, you can either initialize the opentelemetry ruby sdk in function code. To receive the trace, you can define the OTEL-related (e.g. OTEL_EXPORTER_OTLP_ENDPOINT) environmental variable and setup the opentelemetry collector accordingly.

For example:
```ruby
require 'opentelemetry/sdk'
require 'opentelemetry/exporter/otlp'
require 'opentelemetry/instrumentation/all'
OpenTelemetry::SDK.configure do |c|
  c.service_name = '<YOUR_SERVICE_NAME>'
  c.use_all() # enables all instrumentation!
end

def lambda_handler(event:, context:)
  # ... your code
end
```

Sample handler can be found here:
1. [splunk-otel-lambda](https://github.com/signalfx/splunk-otel-lambda/tree/main/ruby)

## Sample App 

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

# curl the lambda function
curl http://127.0.0.1:3000
# Hello 1.3.0% 
```

NOTE:
In `ruby/sample-apps/template.yml`, the OTelLayer -> Properties -> ContentUri is pointing to `ruby/src/layer/`. This is for local testing purpose. If you wish to deploy (e.g. `sam deploy`), please point it to correct location or zip file.

In this sample-apps, we use `src/layer/otel-handler` as default `AWS_LAMBDA_EXEC_WRAPPER`; to change it, please edit in `sample-apps/template.yml`


