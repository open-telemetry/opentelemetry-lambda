# OpenTelemetry Lambda NodeJS

Layer for running NodeJS applications on AWS Lambda with OpenTelemetry. Adding the layer and pointing to it with
the `AWS_LAMBDA_EXEC_WRAPPER` environment variable will initialize OpenTelemetry, enabling tracing with no code change.

To use, add the layer to your function configuration and then set `AWS_LAMBDA_EXEC_WRAPPER` to `/opt/otel-handler`.

[AWS SDK v2 instrumentation](https://github.com/aspecto-io/opentelemetry-ext-js/tree/master/packages/instrumentation-aws-sdk) is also
included and loaded automatically if you use the AWS SDK v2.

## Building

To build the layer and sample applications, in this `nodejs` folder, run

```
npm install
```

This will download all dependencies and compile all code. The layer zip file will be present at `./packages/layer/build/layer.zip`.

## Configuration

Within `packages/layer/src/wrapper.ts`, there are various hooks that can be used to configure tracing.

For each feature (e.g., Tracer, TracerProvider, etc.) to configure, declare the corresponding `configure` function in a file that should be `--require`'d via the Lambda's `NODE_OPTIONS` environment variable.

A full list of configure functions can be found at `packages/layer/src/wrapper.ts`.


e.g., To configure the instrumentations to register:
1. Create a file,`configuration.ts`:
```javascript
declare global {
  function configureInstrumentations(): InstrumentationOption[];
}

global.configureInstrumentations = () => {
  return [new HttpInstrumentation()];
};
```

2. Set the `NODE_OPTIONS` environment variable on your lambda to the _compiled file_ `--require ./path/to/configuration.js`

## Sample applications

Sample applications are provided to show usage of the above layer.

- Application using AWS SDK - shows using the wrapper with an application using AWS SDK without code change.
  - [Using layer built from source](./integration-tests/aws-sdk)
  - [WIP] [Using OTel Public Layer](./sample-apps/aws-sdk)
