module "hello-awssdk-java-wrapper" {
  source = "terraform-aws-modules/lambda/aws"

  function_name = "hello-awssdk-java-wrapper"
  handler       = "io.opentelemetry.lambda.sampleapps.awssdk.AwsSdkRequestHandler::handleRequest"
  runtime       = "java11"

  create_package         = false
  local_existing_package = "${path.module}/../../build/libs/aws-sdk-all.jar"

  memory_size = 384
  timeout     = 20

  layers = [
    var.collector_layer_arn,
    var.javawrapper_layer_arn
  ]

  environment_variables = {
    AWS_LAMBDA_EXEC_WRAPPER = "/opt/otel-proxy-handler"
    OTEL_TRACES_EXPORTER    = "logging"
    OTEL_METRICS_EXPORTER   = "logging"
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

  name                = "hello-java-awssdk-wrapper"
  function_name       = module.hello-awssdk-java-wrapper.this_lambda_function_name
  function_invoke_arn = module.hello-awssdk-java-wrapper.this_lambda_function_invoke_arn
}
