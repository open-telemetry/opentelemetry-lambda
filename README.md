# OpenTelemetry Lambda
![GitHub Java Workflow Status](https://img.shields.io/github/workflow/status/open-telemetry/opentelemetry-lambda/Continuous%20Build%20(Java)/main?label=CI%20%28Java%29&style=for-the-badge)
![GitHub Collector Workflow Status](https://img.shields.io/github/workflow/status/open-telemetry/opentelemetry-lambda/Continuous%20Build%20(Collector)/main?label=CI%20%28Collector%29&style=for-the-badge)
![GitHub NodeJS Workflow Status](https://img.shields.io/github/workflow/status/open-telemetry/opentelemetry-lambda/Continuous%20Build%20(NodeJS)/main?label=CI%20%28NodeJS%29&style=for-the-badge)
![GitHub Terraform Lint Workflow Status](https://img.shields.io/github/workflow/status/open-telemetry/opentelemetry-lambda/Continuous%20Build%20(Terraform)/main?label=CI%20%28Terraform%20Lint%29&style=for-the-badge)
![GitHub Python Pull Request Workflow Status](https://img.shields.io/github/workflow/status/open-telemetry/opentelemetry-lambda/Pull%20Request%20(Python)/main?label=Pull%20Request%20%28Python%29&style=for-the-badge)

## OpenTelemetry Lambda Layers
The OpenTelemetry Lambda Layers provide the OpenTelemetry (OTel) code to export telemetry asynchronously from AWS Lambdas. It does this by embedding a stripped-down version of [OpenTelemetry Collector Contrib](https://github.com/open-telemetry/opentelemetry-collector-contrib) inside an [AWS Lambda Extension Layer](https://aws.amazon.com/blogs/compute/introducing-aws-lambda-extensions-in-preview/).

Some layers include the corresponding OTel language SDK for the Lambda. This allows Lambdas to use OpenTelemetry to send traces and metrics to any configured backend.

## Extension Layer Language Support

* ### [Python Lambda Layer](python/README.md)
* ### [Java Lambda Layer](java/README.md)
* ### [NodeJS Lambda Layer](nodejs/README.md)
* ### [.NET Lambda Layer](dotnet/README.md)


## FAQ
* **What exporters/recievers/processors are included from the OpenTelemetry Collector?**
    > For a full list of currently included components, see [here](https://github.com/open-telemetry/opentelemetry-lambda/blob/main/collector/lambdacomponents/default.go#L18).
* **Is the Lambda layer provided or do I need to build it and distribute it myself?**
    > This repository does not provide pre-build Lambda layers. They must be built manually and saved in your AWS account. This repo has files to facilitate doing that. More information is provided in [the Collector folder's README](collector/README.md).

## Design Proposal
To get a better understanding of the proposed design for the OpenTelemetry Lamda extension, you can the [Design Proposal here.](docs/design_proposal.md)
