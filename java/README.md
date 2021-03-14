# OpenTelemetry Lambda Java

Layers for running Java applications on AWS Lambda with OpenTelemetry.

## Provided layers

Two types of layers are provided

### Java agent

The [AWS OpenTelemetry Java Agent](https://github.com/aws-observability/aws-otel-java-instrumentation)
is bundled into the base of the layer and can be loaded into a Lambda function by specifying the
environment variable `JAVA_TOOL_OPTIONS=-javaagent:/opt/aws-opentelemetry-agent.jar` in your Lambda
configuration. The agent will automatically instrument your application for all supported libraries.
Note, automatic instrumentation has a notable impact on startup time on AWS Lambda and you will
generally need to use this along with provisioned concurrency and warmup requests to serve production
requests without causing timeouts on initial requests while it initializes.

### Wrapper
[OpenTelemetry Lambda Instrumentation](https://github.com/open-telemetry/opentelemetry-java-instrumentation/tree/main/instrumentation/aws-lambda-1.0/library)
and [OpenTelemetry SDK](https://github.com/open-telemetry/opentelemetry-java) are bundled into the
`java/lib` directory to be available on the classpath of the Lambda function. No code change is
needed to instrument the execution of your function, but you will need to change the configuration
to point the `OTEL_INSTRUMENTATION_AWS_LAMBDA_HANDLER` environment variable to your handler function
in the format `package.ClassName::methodName` and set the actual handler of the function to one of

- io.opentelemetry.instrumentation.awslambda.v1_0.TracingRequestWrapper - for wrapping regular handlers (implementing RequestHandler)
- io.opentelemetry.instrumentation.awslambda.v1_0.TracingRequestApiGatewayWrapper - for wrapping regular handlers (implementing RequestHandler) proxied through API Gateway, enabling HTTP context propagation
- io.opentelemetry.instrumentation.awslambda.v1_0.TracingRequestStreamWrapper - for wrapping streaming handlers (implementing RequestStreamHandler), enabling HTTP context propagation for HTTP requests

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

The layer zip file will be present at `./layer-javaagent/build/distributions/aws-opentelemetry-agent-layer.zip`.

To build the wrapper layer, run

```
./gradlew :layer-wrapper:build
```

The layer zip file will be present at `./layer-wrapper/build/distributions/opentelemetry-java-wrapper.zip`.
