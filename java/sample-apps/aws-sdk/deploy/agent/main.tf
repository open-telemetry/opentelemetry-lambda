module "hello-awssdk-javaagent" {
  source = "terraform-aws-modules/lambda/aws"

  function_name = var.name
  handler       = "io.opentelemetry.lambda.sampleapps.awssdk.AwsSdkRequestHandler::handleRequest"
  runtime       = "java11"

  create_package         = false
  local_existing_package = "${path.module}/../../build/libs/aws-sdk-all.jar"

  memory_size = 384
  timeout     = 120
  publish     = true

  layers = [
    var.collector_layer_arn,
    var.javaagent_layer_arn
  ]

  environment_variables = {
    JAVA_TOOL_OPTIONS     = "-javaagent:/opt/opentelemetry-javaagent.jar"
    OTEL_TRACES_EXPORTER  = "logging"
    OTEL_METRICS_EXPORTER = "logging"
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

resource "aws_lambda_alias" "provisioned" {
  name             = "provisioned"
  function_name    = module.hello-awssdk-javaagent.this_lambda_function_name
  function_version = module.hello-awssdk-javaagent.this_lambda_function_version
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
}
