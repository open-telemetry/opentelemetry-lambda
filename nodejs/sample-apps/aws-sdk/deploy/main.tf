module "hello-nodejs" {
  source = "terraform-aws-modules/lambda/aws"

  function_name = "hello-nodejs"
  handler       = "index.handler"
  runtime       = "nodejs14.x"

  create_package         = false
  local_existing_package = "${path.module}/../build/function.zip"

  memory_size = 384
  timeout     = 20

  layers = [
    var.collector_layer_arn,
    var.nodejs_wrapper_layer_arn
  ]

  environment_variables = {
    AWS_LAMBDA_EXEC_WRAPPER = "/opt/otel-handler"
    OTEL_TRACES_EXPORTER    = "logging"
    OTEL_METRICS_EXPORTER   = "logging"
    OTEL_LOG_LEVEL          = "DEBUG"
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
  source = "../../../../utils/terraform/api-gateway-proxy"

  name                = "hello-nodejs"
  function_name       = module.hello-nodejs.this_lambda_function_name
  function_invoke_arn = module.hello-nodejs.this_lambda_function_invoke_arn
}

