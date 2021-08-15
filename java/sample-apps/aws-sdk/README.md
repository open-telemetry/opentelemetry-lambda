# AWS SDK Sample App

This application makes a simple request to `S3.ListBuckets` using the AWS SDK v2. You can find
deployment scripts using Terraform that are configured to deploy this sample app to AWS Lambda and
API Gateway while publishing and using the OpenTelemetry layers.

## Requirements

- Java for building this repository
- [Terraform](https://www.terraform.io/downloads.html)
- AWS credentials, either using environment variables or via the CLI and `aws configure`

First, in the `java` subfolder of this repository, build all the artifacts.

```
./gradlew build
```

Then, decide if you want to try the wrapper or the agent version. Navigate to the appropriate
subfolder of [deploy](./deploy) and deploy with Terraform.

```
terraform init
terraform apply
```

For the agent version, you can optionally enable a pipeline to send Prometheus metrics to Amazon Managed Service for Prometheus (AMP). To enable it, provide the ARN of a lambda layer with a working AMP region and endpoint (see example [configuration](https://aws-otel.github.io/docs/getting-started/prometheus-remote-write-exporter#aws-prometheus-remote-write-exporter)) in a file named `config.yaml` file (similar to the provided [config.yaml](./../../../collector/config.yaml)). After running the `terraform apply`, when prompted, provide the ARN of the layer you created. 

After deployment, a URL which can be used to invoke the function via API Gateway will be displayed. The agent version
tends to take 10-20s for the first request, while the wrapper version tends to take 5-10s. Confirm
that spans are logged in the CloudWatch logs for the function on the AWS Console either for the
[wrapper](https://console.aws.amazon.com/cloudwatch/home?region=us-east-1#logsV2:log-groups/log-group/$252Faws$252Flambda$252Fhello-awssdk-java-wrapper)
or for the [agent](https://console.aws.amazon.com/cloudwatch/home?region=us-east-1#logsV2:log-groups/log-group/$252Faws$252Flambda$252Fhello-awssdk-javaagent).
