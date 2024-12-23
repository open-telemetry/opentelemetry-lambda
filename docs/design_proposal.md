# Design: OpenTelemetry support for Lambda

## 1. Introduction
AWS Lambda (referred to as Lambda herein) is a compute service that lets the user build serverless applications, where the user’s Lambda code runs in a sandbox environment provisioned and managed by Lambda. Since users do not have full control of the environment, running OpenTelemetry SDKs and Collector in Lambda requires additional setup. For example, launching OpenTelemetry Collector and enabling auto-instrumentation in Lambda are not as straightforward as compared to Linux. That is the motivation for why we created this proposal which aims to provide a seamless OpenTelemetry UX for Lambda users.

### Some requirements for this proposal:
* ___[Lambda container](https://aws.amazon.com/blogs/compute/container-reuse-in-lambda/):___
A user’s Lambda function executes in a Linux container (aka sandbox). Lambda provides the user serverless infrastructure by reusing the existing container when there are requests in queue, freezing it when no new requests, thawing the frozen container if new requests come in, and scaling up capacity by expanding the number of containers.

* ___[Lambda Extension](https://docs.aws.amazon.com/lambda/latest/dg/runtimes-extensions-api.html):___
The Lambda Extensions API provides a method for launching a standalone program in a Lambda container. Using the Lambda Extensions API, the OpenTelemetry Collector can run in Lambda.

* ___[Public Lambda layer](https://docs.aws.amazon.com/lambda/latest/dg/configuration-layers.html):___
A Lambda layer is a .zip file archive that contains libraries, a custom runtime, or other dependencies. It is natural to package the Collector in a layer and share with users. For some languages like Python, the UX can be improved by the public SDK layer but not for all languages.

## 2. Technical Challenges
There are additional limitations when running OpenTelemetry in Lambda. Here is a list of some technical challenges we need to address in development.
### 2.1 Lambda freeze
Lambda freezes the container if there are no new requests coming. When it occurs all processes in the container will be frozen. See details in [cgroup freezer](https://www.kernel.org/doc/Documentation/cgroup-v1/freezer-subsystem.txt). If the process uses a timer to flush the buffered data, the timer won’t fire until the container thaws. The interval between the frozen and the thaw state is unpredictable, ranging from seconds to hours. Therefore, the data buffered in the OpenTelemetry SDK and Collector will have unpredictable latency.

In the long run, we hope Lambda provides an IDLE event callback, which should be around 60 seconds. After Lambda realizes it has frozen for a while, it triggers this callback so the user has a chance to flush the data in the buffer. Lambda does not know if there is a new request coming or that it will freeze or thaw. So, an IDLE event is like a wake-up event once the container has been frozen more than 60 seconds.

The short term solution is to flush data out before the Lambda function returns, because it ensures that the Lambda container will not be frozen during an invocation. To achieve this solution: 
1. The SDK side needs to call [force_flush()](https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/trace/sdk.md#forceflush) at the end of Lambda instrumentation. 
2. The Collector needs to remove all processors from config to keep one thread only from the receiver to the exporter. In this way, telemetry data is flushed to the backend synchronously at the end of every invocation. The side effect is slightly increased invocation time.

![Architect](https://user-images.githubusercontent.com/66336933/113330107-3e8b3f80-92d3-11eb-826c-6110773096b5.png)
### 2.2 Memory consumption
For the best performance experience, Lambda provides the memory used metric, which is used as a high-water mark. This metric is equal to cgroup memory.max_usage_in_bytes, including RSS + CACHE. The page CACHE size is approximately equal to the size of dependencies. This means the more dependencies added in Lambda, the higher the memory consumption is in Lambda metrics. Right now the size of OpenTelemetry Collector-contrib is over 100MB and this size can increase as the RSS size (currently at 10MB) grows.

### 2.3 Layer size limit
The layer size limit is a hard limit specifying that the total unzipped size of the function and all layers can't exceed the unzipped deployment package size limit of 250 MB. Due to this constraint, we have to strip down the size of the Collector layer by removing unnecessary components.

With the exception of the Lambda layer, we can keep an eye on [Lambda Container Image](https://docs.aws.amazon.com/lambda/latest/dg/images-create.html), a new feature that can package Lambda function code and dependencies as a container image by Docker CLI. It brings benefits such as higher volume of dependencies, but now it still has limits that cannot add/update Lambda functions and layers once the image is built. It is not possible to provide the finished image to the user directly. Lambda Container Image is still not a replacement of Lambda layer. Furthermore, it does not change the way of memory consumption in Lambda, the larger dependencies need users to apply for higher performance Lambda containers.

### 2.4 Container platform
For programming languages such as Python, building deployment packages in a local development environment is risky. The local environment is often different from the Lambda container environment in meaningful ways. Python wheels with compiled binaries, for example, may work on Mac but fail on Linux (typically with Python gRPC). The solution is building the SDK layer in a Docker Lambda image. Languages which support cross-compiling, such as Java or Go, don't have to be built in docker.
### 2.5 Auto instrumentation
It's important to clarify that auto-instrumentation is not library instrumentation; auto-instrumentation does not require any code changes within the user's application. Refer to the OpenTelemetry doc for [Python](https://opentelemetry-python.readthedocs.io/en/stable/examples/auto-instrumentation/README.html) and [Java](https://opentelemetry.io/docs/java/getting_started/#automatic-instrumentation). Enabling OpenTelemetry auto-instrumentation in Lambda is non-trivial and is language-specific. For more details, refer to [Lambda runtime modification](https://docs.aws.amazon.com/lambda/latest/dg/runtimes-modify.html). Need to mention is right now not all languages support auto-instrumentation in OpenTelemetry.

## 3. Implementation
### 3.1 Three wrappers
We can have three layers of wrapper on a user's Lambda function: ___Lambda function instrumentation___, ___execution wrapper___ and ___Lambda EXEC Wrapper___. Each one can add new functionality by wrapping the previous one, to further improve user experience.
![wrappers](https://user-images.githubusercontent.com/66336933/113330096-3af7b880-92d3-11eb-89bf-580ed8614807.png)
1. ___Lambda function instrumentation___: 
Lambda function instrumentation is the first step of OpenTelemetry Lambda support, which is wrapping the Lambda handler function with beginSpan() and endSpan(), adding OpenTelemetry attributes by following [trace](https://github.com/open-telemetry/semantic-conventions/blob/main/docs/faas/faas-spans.md) and [resource](https://github.com/open-telemetry/semantic-conventions/blob/main/docs/resource/faas.md) FaaS spec. Java and Python implementations are in place, other languages can refer to the code for them.
2. ___Lambda execution wrapper___: 
Lambda execution wrapper uses reflection to dynamically decorate the user’s original Lambda function. It can add functionality to a user's existing Lambda function without changing source code, but the user must change the Lambda handler name and add the original handler path as an environment variable. With OpenTelemetry, the decoration can be anything, such as initializing OpenTelemetry components, calling instrumentation, or even implementing instrumentation directly. If desired, a developer can also merge Lambda function instrumentation and Lambda execution wrapper into one wrapper.
See [Java example](https://github.com/open-telemetry/opentelemetry-java-instrumentation/pull/1471/files) (decorate Lambda instrumentation) and [Python example](https://github.com/open-telemetry/opentelemetry-lambda/blob/main/python/src/otel/otel_sdk/otel_wrapper.py) (initialize OpenTelemetry components for auto-instrumentation).
3. ___Lambda EXEC Wrapper(Lambda native wrapper)___: 
Lambda has a native [Wrapper scripts](https://docs.aws.amazon.com/lambda/latest/dg/runtimes-modify.html#runtime-wrapper) feature which is triggered by environment variable AWS_LAMBDA_EXEC_WRAPPER. The Lambda EXEC Wrapper can couple with an execution wrapper to improve the UX, see [an example in Python](https://github.com/open-telemetry/opentelemetry-lambda/blob/main/python/src/otel/otel_sdk/otel-instrument). By it, users no need to change Lambda handler name and add the original handler path as an environment variable

### 3.2 OpenTelemetry Collector extension Lambda layer
Unless using the SDK in-process exporter, the user must run OpenTelemetry Collector in Lambda sandbox to forward telemetry data to the backend service. Wrapped by Lambda Extensions API, Collector can run in Lambda sandbox independently. We should support that in the OpenTelemetry Lambda Repo.

In previous technical challenges we mentioned two things: Memory consumption and size limit. Both can be solved by a smaller Collector size. At the moment OpenTelemetry Collector contrib binary is over 100MB, all of OpenTelemetry Collector components are built into it. But in a real case, an user may use only a few components. So, when we develop the OpenTelemetry collector extension layer, an ideal solution is to provide a build the Collector layer on demand functionality, by picking up the components used in Collector config. Or, at least provide an interface for dynamically mounting a cherry-pick of Collector components in Lambda extension layer. For the prototype please refer to [aws-observability/aws-otel-lambda](https://github.com/aws-observability/aws-otel-lambda/)

### 3.3 OpenTelemetry SDK layer
As a prototype, Python is implemented in [current Repo](https://github.com/open-telemetry/opentelemetry-lambda/tree/main/python).

We don’t provide SDK public layers for every language. This is required in the following two cases:
Languages with auto-instrumentation support, such as:
* Python
* Java

Scripting languages such as:
* Python
* Node.js
* Ruby

At the moment only Python and Java support auto-instrumentation. Other languages such as Go, .Net support library instrumentation only. For these languages, users must explicitly add OpenTelemetry SDK dependencies and instrumentation code, then re-compile the Lambda application to enable library instrumentation. In such cases we would not provide OpenTelemetry SDK layer because it does not help in compilation and deployment.

Java SDK supports auto-instrumentation through javaagent, which will noticeably increase the program’s start time. In performance testing we see around 40 seconds for Lambda cold start if using OpenTelemetry javaagent. This is a concern in terms of providing a public Lambda layer for Java. We do see some users asking for OpenTelemetry Java auto-instrumentation in Lambda and aren’t concerned about long cold start times. For this reason, we still provide a public Lambda layer for Java and remind users of the impact in documentation.

Even if we don't provide an SDK layer for a language, we still need to provide a sample demonstrating how to use it. The sample would refer to the OpenTelemetry SDK dependencies directly and consume the public Collector Lambda layer. For example, for time-sensitive users we will provide a sample of Java library instrumentation, as a complement of Java(agent based) auto-instrumentation layer.

## 4. CI/CD convention
Each sample application and layer source folder must have a one-click script `run.sh`, to build and deploy OpenTelemetry Lambda application in a uniform manner. It is not only providing a good user experience but also for building a generic CI/CD workflow, github action can simply run tests and publish Lambda layer no matter what the stack tool(Terraform, CloudFormation), or language(Java, Python) the sample application is. 
See an example for publishing [Python layer](../python/src/).