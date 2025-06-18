locals {
  collector_layer_arn = "arn:aws:lambda:${data.aws_region.current.name}:${var.account_id}:layer:opentelemetry-collector-arm64-${var.collector_layer_version}:1"
  sdk_layer_arn       = "arn:aws:lambda:${data.aws_region.current.name}:${var.account_id}:layer:opentelemetry-nodejs-${var.nodejs_layer_version}:1"
}

data "aws_region" "current" {}

module "hello-lambda-function" {
  source  = "terraform-aws-modules/lambda/aws"
  version = "7.21.0"

  architectures = compact([var.architecture])
  function_name = var.name
  handler       = "index.handler"
  runtime       = var.runtime

  create_package         = false
  local_existing_package = "${path.module}/../../build/function.zip"

  memory_size = 384
  timeout     = 20

  layers = compact([
    local.collector_layer_arn,
    local.sdk_layer_arn
  ])

  environment_variables = {
    AWS_LAMBDA_EXEC_WRAPPER     = "/opt/otel-handler"
    OTEL_TRACES_EXPORTER        = "console"
    OTEL_METRICS_EXPORTER       = "console"
    OTEL_LOG_LEVEL              = "DEBUG"
    OTEL_EXPORTER_OTLP_ENDPOINT = "http://localhost:4318/"
    OTEL_TRACES_SAMPLER         = "always_on"
  }

  tracing_mode = var.tracing_mode
}

module "api-gateway" {
  source = "../../../../../utils/terraform/api-gateway-proxy"

  name                = var.name
  function_name       = module.hello-lambda-function.lambda_function_name
  function_invoke_arn = module.hello-lambda-function.lambda_function_invoke_arn
  enable_xray_tracing = var.tracing_mode == "Active"
}
