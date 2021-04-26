module "hello-lambda-function" {
  source = "terraform-aws-modules/lambda/aws"

  function_name = var.name
  handler       = "io.opentelemetry.lambda.sampleapps.awssdk.AwsSdkRequestHandler::handleRequest"
  runtime       = "java11"

  create_package         = false
  local_existing_package = "${path.module}/../../build/libs/aws-sdk-all.jar"

  memory_size = 512
  timeout     = 120
  publish     = true

  layers = compact([
    var.collector_layer_arn,
    var.sdk_layer_arn
  ])

  environment_variables = {
    AWS_LAMBDA_EXEC_WRAPPER = "/opt/otel-handler"
  }

  tracing_mode = var.tracing_mode

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

resource "aws_lambda_alias" "provisioned" {
  name             = "provisioned"
  function_name    = module.hello-lambda-function.lambda_function_name
  function_version = module.hello-lambda-function.lambda_function_version
}

resource "aws_lambda_provisioned_concurrency_config" "lambda_api" {
  function_name                     = aws_lambda_alias.provisioned.function_name
  provisioned_concurrent_executions = 2
  qualifier                         = aws_lambda_alias.provisioned.name
}

module "api-gateway" {
  source = "../../../../../utils/terraform/api-gateway-proxy"

  name                = var.name
  function_name       = aws_lambda_alias.provisioned.function_name
  function_qualifier  = aws_lambda_alias.provisioned.name
  function_invoke_arn = aws_lambda_alias.provisioned.invoke_arn
  enable_xray_tracing = var.tracing_mode == "Active"
}
