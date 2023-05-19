# OpenTelemetry Lambda Python

Scripts and files used to build AWS Lambda Layers for running OpenTelemetry on AWS Lambda for Python.

### Sample App 

1. Install 
* [SAM CLI](https://docs.aws.amazon.com/serverless-application-model/latest/developerguide/serverless-sam-cli-install.html)
* [AWS CLI](https://docs.aws.amazon.com/cli/latest/userguide/install-cliv2.html)
* [Go](https://go.dev/doc/install)
* [Docker](https://docs.docker.com/get-docker)
3. Run aws configure to [set aws credential(with administrator permissions)](https://docs.aws.amazon.com/serverless-application-model/latest/developerguide/serverless-sam-cli-install-mac.html#serverless-sam-cli-install-mac-iam-permissions) and default region.
4. Download a local copy of this repository from Github.
5. `cd python/sample-apps`
6. If you just want to create a zip file with the OpenTelemetry Python AWS Lambda layer, then use the `-b true` option: `bash run.sh -n <LAYER_NAME_HERE> -b true`
7. If you want to create the layer and automatically publish it, use no options: `bash run.sh`
