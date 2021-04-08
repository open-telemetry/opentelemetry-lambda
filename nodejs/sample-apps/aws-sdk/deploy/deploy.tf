resource "aws_lambda_layer_version" "opentelemetry_nodejs_wrapper" {
  layer_name = "opentelemetry-nodejs-wrapper"
  filename = "../../../packages/layer/build/layer.zip"
  compatible_runtimes = ["nodejs10.x", "nodejs12.x", "nodejs14.x"]
  license_info = "Apache-2.0"
  source_code_hash = filebase64sha256("../../../packages/layer/build/layer.zip")
}

module "hello-nodejs" {
  source = "terraform-aws-modules/lambda/aws"

  function_name = "hello-nodejs"
  handler       = "index.handler"
  runtime = "nodejs14.x"

  create_package         = false
  local_existing_package = "../build/function.zip"

  memory_size = 384
  timeout = 20

  layers = [
    aws_lambda_layer_version.opentelemetry_nodejs_wrapper.arn
  ]

  environment_variables = {
    AWS_LAMBDA_EXEC_WRAPPER = "/opt/otel-handler"
    OTEL_TRACES_EXPORTER = "logging"
    OTEL_METRICS_EXPORTER = "logging"
    OTEL_LOG_LEVEL = "DEBUG"
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

resource "aws_api_gateway_rest_api" "lambda_api_nodejs" {
  name = "hello-lambda-nodejs"
}

resource "aws_api_gateway_resource" "lambda_api_proxy_nodejs" {
  rest_api_id = aws_api_gateway_rest_api.lambda_api_nodejs.id
  parent_id = aws_api_gateway_rest_api.lambda_api_nodejs.root_resource_id
  path_part = "{proxy+}"
}

resource "aws_api_gateway_method" "lambda_api_proxy_nodejs" {
  rest_api_id = aws_api_gateway_rest_api.lambda_api_nodejs.id
  resource_id = aws_api_gateway_resource.lambda_api_proxy_nodejs.id
  http_method = "ANY"
  authorization = "NONE"
}

resource "aws_api_gateway_integration" "lambda_api_nodejs" {
  rest_api_id = aws_api_gateway_rest_api.lambda_api_nodejs.id
  resource_id = aws_api_gateway_method.lambda_api_proxy_nodejs.resource_id
  http_method = aws_api_gateway_method.lambda_api_proxy_nodejs.http_method

  integration_http_method = "POST"
  type = "AWS_PROXY"
  uri = module.hello-nodejs.this_lambda_function_invoke_arn
}

resource "aws_api_gateway_method" "lambda_api_proxy_root_nodejs" {
  rest_api_id = aws_api_gateway_rest_api.lambda_api_nodejs.id
  resource_id = aws_api_gateway_rest_api.lambda_api_nodejs.root_resource_id
  http_method = "ANY"
  authorization = "NONE"
}

resource "aws_api_gateway_integration" "lambda_api_root_nodejs" {
  rest_api_id = aws_api_gateway_rest_api.lambda_api_nodejs.id
  resource_id = aws_api_gateway_method.lambda_api_proxy_root_nodejs.resource_id
  http_method = aws_api_gateway_method.lambda_api_proxy_root_nodejs.http_method

  integration_http_method = "POST"
  type = "AWS_PROXY"
  uri = module.hello-nodejs.this_lambda_function_invoke_arn
}

resource "aws_api_gateway_deployment" "lambda_api_nodejs" {
  depends_on = [
    aws_api_gateway_integration.lambda_api_nodejs,
    aws_api_gateway_integration.lambda_api_root_nodejs,
  ]

  rest_api_id = aws_api_gateway_rest_api.lambda_api_nodejs.id
}

resource "aws_api_gateway_stage" "test_nodejs" {
  stage_name = "default"
  rest_api_id = aws_api_gateway_rest_api.lambda_api_nodejs.id
  deployment_id = aws_api_gateway_deployment.lambda_api_nodejs.id
}

resource "aws_lambda_permission" "lambda_api_allow_gateway_nodejs" {
  action = "lambda:InvokeFunction"
  function_name = module.hello-nodejs.this_lambda_function_name
  principal = "apigateway.amazonaws.com"
  source_arn = "${aws_api_gateway_rest_api.lambda_api_nodejs.execution_arn}/*/*"
}

output "lambda_api_gateway_nodejs_url" {
  value = aws_api_gateway_stage.test_nodejs.invoke_url
}
