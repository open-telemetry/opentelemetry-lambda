# OpenTelemetry Lambda Java

Layers for running Java applications on AWS Lambda with OpenTelemetry.

## Prerequisites 
- Supports functions running on Java 11 corretto runtime only. 

## Provided layers

Two types of layers are provided

### Java agent

The [OpenTelemetry Java Agent](https://github.com/open-telemetry/opentelemetry-java-instrumentation)
is bundled into the base of the layer and can be loaded into a Lambda function by specifying the
`AWS_LAMBDA_EXEC_WRAPPER=/opt/otel-handler` in your Lambda configuration. The agent will be automatically
loaded and instrument your application for all supported libraries.

Note, automatic instrumentation has a notable impact on startup time on AWS Lambda and you will
generally need to use this along with provisioned concurrency and warmup requests to serve production
requests without causing timeouts on initial requests while it initializes.

### Wrapper
[OpenTelemetry Lambda Instrumentation](https://github.com/open-telemetry/opentelemetry-java-instrumentation/tree/main/instrumentation/aws-lambda-1.0/library)
and [OpenTelemetry SDK](https://github.com/open-telemetry/opentelemetry-java) are bundled into the
`java/lib` directory to be available on the classpath of the Lambda function. No code change is
needed to instrument the execution of your function, but you will need to set the `AWS_LAMBDA_EXEC_WRAPPER`
environment variable pointing to the appropriate wrapper for the type of handler.

- `/opt/otel-handler` - for wrapping regular handlers (implementing RequestHandler)
- `/opt/otel-proxy-handler` - for wrapping regular handlers (implementing RequestHandler) proxied through API Gateway, enabling HTTP context propagation
- `/opt/otel-stream-handler` - for wrapping streaming handlers (implementing RequestStreamHandler), enabling HTTP context propagation for HTTP requests

[AWS SDK instrumentation](https://github.com/open-telemetry/opentelemetry-java-instrumentation/tree/main/instrumentation/aws-sdk/aws-sdk-2.2/library) is also
included and loaded automatically if you use the AWS SDK.

For any other library, such as OkHttp, you will need to include the corresponding library instrumentation
from the [instrumentation project](https://github.com/open-telemetry/opentelemetry-java-instrumentation) and
modify your code to initialize it in your function.

## Building

To build the Java Agent layer, run

```
./gradlew :layer-javaagent:build
```

The layer zip file will be present at `./layer-javaagent/build/distributions/opentelemetry-javaagent-layer.zip`.

To build the wrapper layer, run

```
./gradlew :layer-wrapper:build
```

The layer zip file will be present at `./layer-wrapper/build/distributions/opentelemetry-java-wrapper.zip`.

## Sample applications

Sample applications are provided to show usage the above layers.

- [Application using AWS SDK](./sample-apps/aws-sdk) - shows how both the wrapper and agent can be used
with an application using AWS SDK without code change.

- [Application using OkHttp](./sample-apps/okhttp) - shows the manual initialization of OkHttp
library instrumentation for use with the wrapper. The agent would be usable without such a code change
at the expense of the cold start overhead it introduces.
