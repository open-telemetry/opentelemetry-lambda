# OpenTelemetry Lambda

![GitHub Java Workflow Status](https://img.shields.io/github/actions/workflow/status/open-telemetry/opentelemetry-lambda/ci-java.yml?branch%3Amain&label=CI%20%28Java%29&style=for-the-badge)
![GitHub Collector Workflow Status](https://img.shields.io/github/actions/workflow/status/open-telemetry/opentelemetry-lambda/ci-collector.yml?branch%3Amain&label=CI%20%28Collector%29&style=for-the-badge)
![GitHub NodeJS Workflow Status](https://img.shields.io/github/actions/workflow/status/open-telemetry/opentelemetry-lambda/ci-nodejs.yml?branch%3Amain&label=CI%20%28NodeJS%29&style=for-the-badge)
![GitHub Terraform Lint Workflow Status](https://img.shields.io/github/actions/workflow/status/open-telemetry/opentelemetry-lambda/ci-terraform.yml?branch%3Amain&label=CI%20%28Terraform%20Lint%29&style=for-the-badge)
![GitHub Python Pull Request Workflow Status](https://img.shields.io/github/actions/workflow/status/open-telemetry/opentelemetry-lambda/ci-python.yml?branch%3Amain&label=Pull%20Request%20%28Python%29&style=for-the-badge)

## OpenTelemetry Lambda Layers

The OpenTelemetry Lambda Layers provide the OpenTelemetry (OTel) code to export telemetry asynchronously from AWS Lambda functions. It does this by embedding a stripped-down version of [OpenTelemetry Collector Contrib](https://github.com/open-telemetry/opentelemetry-collector-contrib) inside an [AWS Lambda Extension Layer](https://aws.amazon.com/blogs/compute/introducing-aws-lambda-extensions-in-preview/).

Some layers include the corresponding OTel language SDK for the Lambda. This allows Lambda functions to use OpenTelemetry to send traces and metrics to any configured backend.

## Extension Layer Language Support

* ### [Python + Collector Lambda Layer](python/README.md)
* ### [Java + Collector Lambda Layer](java/README.md)
* ### [NodeJS + Collector Lambda Layer](nodejs/README.md)
* ### [.NET + Collector Lambda Layer](dotnet/README.md)
* ### [Ruby + Collector Lambda Layer](ruby/README.md)
* ### [Collector Lambda Layer](collector/README.md)

## FAQ

* **What exporters/receivers/processors are included from the OpenTelemetry Collector?**
    > You can check out [the stripped-down collector's imports](https://github.com/open-telemetry/opentelemetry-lambda/blob/main/collector/lambdacomponents/default.go#L18) in this repository for a full list of currently included components.
* **Is the Lambda layer provided or do I need to build it and distribute it myself?**
    > This repository provides pre-built Lambda layers, their ARNs are available in the [Releases](https://github.com/open-telemetry/opentelemetry-lambda/releases). You can also build the layers manually and publish them in your AWS account. This repo has files to facilitate doing that. More information is provided in [the Collector folder's README](collector/README.md).

## Design Proposal

To get a better understanding of the proposed design for the OpenTelemetry Lambda extension, you can see the [Design Proposal here.](docs/design_proposal.md)

## Features

The following is a list of features provided by the OpenTelemetry layers.

### OpenTelemetry collector

The layer includes the OpenTelemetry Collector as a Lambda extension.

### Custom context propagation carrier extraction

Context can be propagated through various mechanisms (e.g. http headers (APIGW), message attributes (SQS), ...). In some cases, it may be required to pass a custom context propagation extractor in Lambda through configuration, this feature allows this through Lambda instrumentation configuration.

### X-Ray Env Var Span Link

This links a context extracted from the Lambda runtime environment to the instrumentation-generated span rather than disabling that context extraction entirely.

### Semantic conventions

The Lambda language implementation follows the semantic conventions specified in the OpenTelemetry Specification.

### Auto instrumentation

The Lambda layer includes support for automatically instrumentation code via the use of instrumentation libraries.

### Flush TracerProvider

The Lambda instrumentation will flush the `TracerProvider` at the end of an invocation.

### Flush MeterProvider

The Lambda instrumentation will flush the `MeterProvider` at the end of an invocation.

### Support matrix

The table below captures the state of various features and their levels of support different runtimes.

| Feature                    | Node | Python | Java | .NET | Go   | Ruby |
| -------------------------- | :--: | :----: | :--: | :--: | :--: | :--: |
| OpenTelemetry collector    |  +   |  +     |  +   |  +   |  +   |  +   |
| Custom context propagation |  +   |  -     |  -   |  -   | N/A  |  +   |
| X-Ray Env Var Span Link    |  -   |  -     |  -   |  -   | N/A  |  -   |
| Semantic Conventions^      |      |  +     |  +   |  +   | N/A  |  +   |
| - Trace General^<sup>[1]</sup>           |  +   |        |  +   |  +   | N/A  |   +  |
| - Trace Incoming^<sup>[2]</sup>          |  -   |        |  -   |  +   | N/A  |   -  |
| - Trace Outgoing^<sup>[3]</sup>          |  +   |        |  -   |  +   | N/A  |   +  |
| - Metrics^<sup>[4]</sup>                 |  -   |        |  -   |  -   | N/A  |   -  |
| Auto instrumentation       |  +   |   +    |  +   |  -   | N/A  |   +  |
| Flush TracerProvider       |  +   |   +    |      |  +   |  +   |   +  |
| Flush MeterProvider        |  +   |   +    |      |      |      |   -  |

#### Legend

* `+` is supported
* `-` not supported
* `^` subject to change depending on spec updates
* `N/A` not applicable to the particular language
* blank cell means the status of the feature is not known.

The following are runtimes which are no longer or not yet supported by this repository:

* Node.js 12 - not [officially supported](https://github.com/open-telemetry/opentelemetry-js#supported-runtimes) by OpenTelemetry JS

[1]: https://github.com/open-telemetry/semantic-conventions/blob/main/docs/faas/faas-spans.md#general-attributes
[2]: https://github.com/open-telemetry/semantic-conventions/blob/main/docs/faas/faas-spans.md#incoming-invocations
[3]: https://github.com/open-telemetry/semantic-conventions/blob/main/docs/faas/faas-spans.md#outgoing-invocations
[4]: https://github.com/open-telemetry/semantic-conventions/blob/main/docs/faas/faas-metrics.md

## Contributing

See the [Contributing Guide](CONTRIBUTING.md) for details.

Here is a list of community roles with current and previous members:

- Approvers ([@open-telemetry/lambda-extension-approvers](https://github.com/orgs/open-telemetry/teams/lambda-extension-approvers)):

  - [Nathan Slaughter](https://github.com/nslaughter), Lightstep
  - [Serkan Özal](https://github.com/serkan-ozal), Catchpoint

- Emeritus Approvers:

  - [Lei Wang](https://github.com/wangzlei)
  - [Nathaniel Ruiz Nowell](https://github.com/NathanielRN)
  - [Tristan Sloughter](https://github.com/tsloughter)

- Maintainers ([@open-telemetry/lambda-extension-maintainers](https://github.com/orgs/open-telemetry/teams/lambda-extension-maintainers)):

  - [Raphael Philipe Mendes da Silva](https://github.com/rapphil), AWS
  - [Tyler Benson](https://github.com/tylerbenson), Lightstep

- Emeritus Maintainers:

  - [Alex Boten](https://github.com/codeboten)
  - [Anthony Mirabella](https://github.com/Aneurysm9)

Learn more about roles in the [community repository](https://github.com/open-telemetry/community/blob/main/community-membership.md).
