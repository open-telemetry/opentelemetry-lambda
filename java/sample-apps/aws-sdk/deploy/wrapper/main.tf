module "hello-awssdk-java-wrapper" {
  source = "terraform-aws-modules/lambda/aws"

  function_name = var.name
  handler       = "io.opentelemetry.lambda.sampleapps.awssdk.AwsSdkRequestHandler::handleRequest"
  runtime       = "java11"

  create_package         = false
  local_existing_package = "${path.module}/../../build/libs/aws-sdk-all.jar"

  memory_size = 384
  timeout     = 20

  layers = concat(
    [
      var.collector_layer_arn,
      var.java_wrapper_layer_arn
    ],
  var.enable_lambda_insights ? ["arn:aws:lambda:us-east-1:580247275435:layer:LambdaInsightsExtension:14"] : [])

  environment_variables = {
    AWS_LAMBDA_EXEC_WRAPPER = "/opt/otel-proxy-handler"
  }

  attach_policy_statements = true
  policy_statements = {
    s3 = {
      effect = "Allow"
      actions = [
        "s3:ListAllMyBuckets"
      ]
      resources = [
        "*"
      ]
    }
  }
}

module "api-gateway" {
  source = "../../../../../utils/terraform/api-gateway-proxy"

  name                = var.name
  function_name       = module.hello-awssdk-java-wrapper.this_lambda_function_name
  function_invoke_arn = module.hello-awssdk-java-wrapper.this_lambda_function_invoke_arn
}
