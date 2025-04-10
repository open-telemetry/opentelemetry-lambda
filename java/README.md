# OpenTelemetry Lambda Java

Layers for running Java applications on AWS Lambda with OpenTelemetry.

## Prerequisites

- Supports Lambda functions using Java 8 (java8.al2) and later
- The deprecated runtime Java 8 (java8) is not supported as it does not support [exec wrappers](https://docs.aws.amazon.com/lambda/latest/dg/runtimes-modify.html#runtime-wrapper).

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

#### Fast startup for Java agent

Fast startup mode is disabled by default but can be enabled by specifying the `OTEL_JAVA_AGENT_FAST_STARTUP_ENABLED=true`
in your Lambda configuration.

When fast startup mode is enabled, **JIT** (Just-In-Time) **Tiered compilation** is configured to stop at level 1 
and bytecode verification is disabled. So, the JVM uses the **C1** compiler which is optimized for fast start-up time. 
This compiler (**C1**) quickly produces optimized native code 
but it does not generate any profiling data and never uses the **C2** compiler 
which optimized for the best overall performance but uses more memory and takes a longer time to achieve it.
Therefore, this option is not enabled by default and needs to be enabled by the user explicitly 
by taking care of the behavioural change mentioned above.

For more information about the idea behind this optimization, you can check the following resources:
- https://aws.amazon.com/tr/blogs/compute/optimizing-aws-lambda-function-performance-for-java/
- https://aws.amazon.com/tr/blogs/compute/increasing-performance-of-java-aws-lambda-functions-using-tiered-compilation/

### Wrapper

[OpenTelemetry Lambda Instrumentation](https://github.com/open-telemetry/opentelemetry-java-instrumentation/tree/main/instrumentation/aws-lambda)
and [OpenTelemetry SDK](https://github.com/open-telemetry/opentelemetry-java) are bundled into the
`java/lib` directory to be available on the classpath of the Lambda function. No code change is
needed to instrument the execution of your function, but you will need to set the `AWS_LAMBDA_EXEC_WRAPPER`
environment variable pointing to the appropriate wrapper for the type of handler.

- `/opt/otel-handler` - for wrapping regular handlers (implementing RequestHandler)
- `/opt/otel-sqs-handler` - for wrapping SQS-triggered function handlers (implementing RequestHandler)
- `/opt/otel-proxy-handler` - for wrapping regular handlers (implementing RequestHandler) proxied through API Gateway, enabling HTTP context propagation
- `/opt/otel-stream-handler` - for wrapping streaming handlers (implementing RequestStreamHandler), enabling HTTP context propagation for HTTP requests

[AWS SDK instrumentation](https://github.com/open-telemetry/opentelemetry-java-instrumentation/tree/main/instrumentation/aws-sdk/aws-sdk-2.2/library) is also
included and loaded automatically if you use the AWS SDK.

For any other library, such as OkHttp, you will need to include the corresponding library instrumentation
from the [instrumentation project](https://github.com/open-telemetry/opentelemetry-java-instrumentation) and
modify your code to initialize it in your function.

## Configuring Context Propagators

### If you emit your traces to AWS X-Ray (instead of a third-party service) and have enabled X-Ray Active Tracing
Please use the following environment variable for your lambda:
`OTEL_PROPAGATORS=tracecontext,baggage,xray-lambda`

### If you emit your traces to another system besides AWS X-Ray (Default Setting)
The following propagators are configured by default, but you can use this environment variable to customize it:
`OTEL_PROPAGATORS=tracecontext,baggage,xray`


For more information please read: https://opentelemetry.io/docs/specs/semconv/faas/aws-lambda

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

The layer zip file will be present at `./layer-wrapper/build/distributions/opentelemetry-javawrapper-layer.zip`.

## Sample applications

Sample applications are provided to show usage the above layers.

- [Application using AWS SDK](./sample-apps/aws-sdk) - shows how both the wrapper and agent can be used
  with an application using AWS SDK without code change.

- [Application using OkHttp](./sample-apps/okhttp) - shows the manual initialization of OkHttp
  library instrumentation for use with the wrapper. The agent would be usable without such a code change
  at the expense of the cold start overhead it introduces.
