# OkHttp Sample App

This application makes a simple request to `https://aws.amazon.com` using OkHttp. You can find
deployment scripts using Terraform that are configured to deploy this sample app to AWS Lambda and
API Gateway while publishing and using the OpenTelemetry layers.

Notice that we initialize the library instrumentation for OkHttp provided by the
`io.opentelemetry.instrumentation:opentelemetry-okhttp-3.0` artifact using our own code,

```java
OkHttpClient client = 
    new OkHttpClient.Builder()
      .addInterceptor(OkHttpTracing.create(GlobalOpenTelemetry.get()).newInterceptor())
      .build();
```

This is to allow the application to be deployed with the wrapper and still trace the OkHttp client.
If using the agent, this is unnecessary.

## Requirements

- Java for building this repository
- [Terraform](https://www.terraform.io/downloads.html)
- AWS credentials, either using environment variables or via the CLI and `aws configure`

First, in the `java` subfolder of this repository, build all the artifacts.

```
./gradlew build
```

Then, navigate to [deploy/wrapper](./deploy/wrapper) and deploy with Terraform.

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

After deployment, a URL which can be used to invoke the function via API Gateway will be displayed. As it uses the 
wrapper, it tends to take 5-10s. Confirm that spans are logged in the CloudWatch logs for the function on the AWS Console for the
[wrapper](https://console.aws.amazon.com/cloudwatch/home?region=us-east-1#logsV2:log-groups/log-group/$252Faws$252Flambda$252Fhello-awssdk-java-wrapper).n

Note that this example cannot currently be used with the agent because it does not behave correctly
with applications that manually initialize library instrumentation. This issue will be fixed in a
future version.
