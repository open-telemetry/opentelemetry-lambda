# Amazon SQS Sample App

This application includes a Lambda with an SQS queue as an event source. You can find
deployment scripts using Terraform that are configured to deploy this sample app to AWS Lambda while publishing and using the OpenTelemetry layers.

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

Use the following command to configure runtime and architecture

```
TF_VAR_architecture=x86_64 \
TF_VAR_runtime=java11 \
terraform apply -auto-approve
```

For the agent version, to change the configuration of the OpenTelemetry collector, you can provide the ARN of a Lambda layer with a custom collector configuration in a file named `config.yaml` when prompted after running the `terraform apply` command.

After deployment, login to AWS and test the Lambda function using the predefined SQS test payload. The agent version tends to take 10-20s for the first request, while the wrapper version tends to take 5-10s. Confirm
that spans are logged in the CloudWatch logs for the function on the AWS Console either for the
[wrapper](https://console.aws.amazon.com/cloudwatch/home?region=us-east-1#logsV2:log-groups/log-group/$252Faws$252Flambda$252Fhello-awssdk-java-wrapper)
or for the [agent](https://console.aws.amazon.com/cloudwatch/home?region=us-east-1#logsV2:log-groups/log-group/$252Faws$252Flambda$252Fhello-awssdk-javaagent).

If you already have an SQS queue to test with, uncomment the commented sections at the bottom of the `main.tf` and `variables.tf` files and then provide the ARN for your queue when running `terraform apply`. You will then need to publish a message to the queue to test the Lambda function.
