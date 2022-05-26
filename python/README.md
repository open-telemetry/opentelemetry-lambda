# OpenTelemetry Lambda Python

Scripts and files used to build AWS Lambda Layers for running OpenTelemetry on AWS Lambda for Python.

## Wrapper

The wrapper is currently based on `OpenTelemetry Python` release `1.12.0rc1-0.31b0`.

## Sample App 

1. Install [SAM CLI](https://docs.aws.amazon.com/serverless-application-model/latest/developerguide/serverless-sam-cli-install.html) and [AWS CLI](https://docs.aws.amazon.com/cli/latest/userguide/install-cliv2.html).
2. Run aws configure to [set aws credential(with administrator permissions)](https://docs.aws.amazon.com/serverless-application-model/latest/developerguide/serverless-sam-cli-install-mac.html#serverless-sam-cli-install-mac-iam-permissions) and default region.
3. Download a local copy of this repository from Github.
4. cd python/sample-apps && ./run.sh
