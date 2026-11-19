# OpenTelemetry Lambda NodeJS

Layer for running NodeJS applications on AWS Lambda with OpenTelemetry. Adding the layer and pointing to it with
the `AWS_LAMBDA_EXEC_WRAPPER` environment variable will initialize OpenTelemetry, enabling tracing with no code change.

To use, add the layer to your function configuration and then set `AWS_LAMBDA_EXEC_WRAPPER` to `/opt/otel-handler`.

## Configuration hooks

The following hooks are provided as a way to reconfigure the layer from within a function:

| Name | Config Type | Brief |
| —— | —— | —— |
| configureAwsInstrumentation | [AwsSdkInstrumentationConfig][https://github.com/open-telemetry/opentelemetry-js-contrib/blob/main/packages/instrumentation-aws-sdk/src/types.ts) | |
| configureLambdaInstrumentation | [AwsLambdaInstrumentationConfig](https://github.com/open-telemetry/opentelemetry-js-contrib/blob/main/packages/instrumentation-aws-lambda/src/types.ts) | |
| configureInstrumentation | | |
| configureSdk | [NodeSDKConfiguration](https://github.com/open-telemetry/opentelemetry-js/blob/main/experimental/packages/opentelemetry-sdk-node/src/types.ts) | provides full access to define the configuration used to start otel. |

The following hooks are deprecated & will be removed in a future release:

- configureLogger
- configureMeter
- configureSdkRegistration
- configureTracer

To configure the sdk use the configureSdk hook, which provides access to all the options available to configure the sdk on node.

## Configuring auto instrumentation

[AWS SDK v3 instrumentation](https://github.com/open-telemetry/opentelemetry-js-contrib/tree/main/packages/instrumentation-aws-sdk)
is included and loaded automatically by default.
A subset of instrumentations from the [OTEL auto-instrumentations-node metapackage](https://github.com/open-telemetry/opentelemetry-js-contrib/tree/main/packages/auto-instrumentations-node)
are also included.

Following instrumentations from the metapackage are included:
- `amqplib`
- `bunyan`
- `cassandra-driver`
- `connect`
- `dataloader`
- `dns` *- default*
- `express` *- default*
- `fs`
- `graphql` *- default*
- `grpc` *- default*
- `hapi` *- default*
- `http` *- default*
- `ioredis` *- default*
- `kafkajs`
- `knex`
- `koa` *- default*
- `memcached`
- `mongodb` *- default*
- `mongoose`
- `mysql` *- default*
- `mysql2`
- `nestjs-core`
- `net` *- default*
- `pg` *- default*
- `pino`
- `redis` *- default*
- `restify`
- `socket.io`
- `undici`
- `winston`

Instrumentations annotated with "*- default*" are loaded by default.

To only load specific instrumentations, specify the `OTEL_NODE_ENABLED_INSTRUMENTATIONS` environment variable in the lambda configuration.
This disables all the defaults, and only enables the ones you specify. Selectively disabling instrumentations from the defaults is also possible with the `OTEL_NODE_DISABLED_INSTRUMENTATIONS` environment variable.

The environment variables should be set to a comma-separated list of the instrumentation package names without the
`@opentelemetry/instrumentation-` prefix.

For example, to enable only `@opentelemetry/instrumentation-http` and `@opentelemetry/instrumentation-undici`:
```shell
OTEL_NODE_ENABLED_INSTRUMENTATIONS="http,undici"
```
To disable only `@opentelemetry/instrumentation-net`:
```shell
OTEL_NODE_DISABLED_INSTRUMENTATIONS="net"
```

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
  - [WIP] [Using OTel Public Layer](./sample-apps/aws-sdk)
