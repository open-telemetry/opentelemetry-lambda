# OpenTelemetry Lambda NodeJS

Layer for running NodeJS applications on AWS Lambda with OpenTelemetry. Adding the layer and pointing to it with
the `AWS_LAMBDA_EXEC_WRAPPER` environment variable will initialize OpenTelemetry, enabling tracing with no code change.

To use, add the layer to your function configuration and then set `AWS_LAMBDA_EXEC_WRAPPER` to `/opt/otel-handler`.

[AWS SDK v2 instrumentation](https://github.com/aspecto-io/opentelemetry-ext-js/tree/master/packages/instrumentation-aws-sdk) is also
included and loaded automatically if you use the AWS SDK v2.

## Supported Runtimes

| Platform Version    | Supported                                     |
| ------------------- | --------------------------------------------- |
| `nodejs22.x`        | :heavy_check_mark:                            |
| `nodejs20.x`        | :heavy_check_mark:                            |
| `nodejs18.x`        | :heavy_check_mark:                            |
| Older Node Versions | See [Node Support](#node-support)             |

### Node Support

Only Node.js Active or Maintenance LTS versions are supported.

## Building

To build the layer and sample applications in this `nodejs` folder:

First install dependencies:

```
npm install
```

Then build the project:

```
npm run build
```

You'll find the generated layer zip file at `./packages/layer/build/layer.zip`.

## Sample applications

Sample applications are provided to show usage of the above layer.

- Application using AWS SDK - shows using the wrapper with an application using AWS SDK without code change.
  - [Using layer built from source](./integration-tests/aws-sdk)
  - [WIP] [Using OTel Public Layer](./sample-apps/aws-sdk)
