# AWS SDK Sample Application

This sample application demonstrates usage of the AWS SDK. To try it out, make sure the collector and nodejs Lambda
layers are built.

In [collector](../../../collector), run `make publish-layer` and note the ARN of the published collector layer.
In [nodejs](../../), run `npm install` then upload the built layer to your AWS account. Note the ARN of the published SDK layer.

Then, navigate to the `deploy` directory and run `terraform init` and `terraform apply`. Input the ARNs of the collector and SDK layers when prompted. You may also need to set the `AWS_DEFAULT_REGION` environment variable to your preferred AWS region.

The lambda function will be initialized and the URL for an API Gateway invoking the Lambda
will be displayed at the end. Send a request to the URL in a browser or using curl to execute the function. Then,
navigate to the function's logs [here](https://console.aws.amazon.com/cloudwatch/home?region=us-east-1#logStream:group=%252Faws%252Flambda%252Fhello-nodejs).
You will see a log stream with an event time corresponding to when you issued the request - open it and you can find
information about the exported spans in the log stream.
