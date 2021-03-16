resource "aws_lambda_layer_version" "opentelemetry_javaagent" {
  layer_name = "opentelemetry-javaagent"
  filename = "../../../../layer-javaagent/build/distributions/opentelemetry-javaagent-layer.zip"
  compatible_runtimes = ["java8", "java8.al2", "java11"]
  license_info = "Apache-2.0"
  source_code_hash = filebase64sha256("../../../../layer-javaagent/build/distributions/opentelemetry-javaagent-layer.zip")
}

module "hello-awssdk-javaagent" {
  source = "terraform-aws-modules/lambda/aws"

  function_name = "hello-awssdk-javaagent"
  handler       = "io.opentelemetry.lambda.sampleapps.awssdk.AwsSdkRequestHandler::handleRequest"
  runtime = "java11"

  create_package         = false
  local_existing_package = "../../build/libs/aws-sdk-all.jar"

  memory_size = 384
  timeout = 120
  publish = true

  layers = [
    aws_lambda_layer_version.opentelemetry_javaagent.arn
  ]

  environment_variables = {
    JAVA_TOOL_OPTIONS = "-javaagent:/opt/opentelemetry-javaagent.jar"
    OTEL_TRACES_EXPORTER = "logging"
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
  name = "provisioned"
  function_name = module.hello-awssdk-javaagent.this_lambda_function_name
  function_version = module.hello-awssdk-javaagent.this_lambda_function_version
}

resource "aws_lambda_provisioned_concurrency_config" "lambda_api" {
  function_name = aws_lambda_alias.provisioned.function_name
  provisioned_concurrent_executions = 2
  qualifier = aws_lambda_alias.provisioned.name
}

resource "aws_api_gateway_rest_api" "lambda_api_javaagent" {
  name = "hello-lambda-awssdk-javaagent"
}

resource "aws_api_gateway_resource" "lambda_api_proxy_javaagent" {
  rest_api_id = aws_api_gateway_rest_api.lambda_api_javaagent.id
  parent_id = aws_api_gateway_rest_api.lambda_api_javaagent.root_resource_id
  path_part = "{proxy+}"
}

resource "aws_api_gateway_method" "lambda_api_proxy_javaagent" {
  rest_api_id = aws_api_gateway_rest_api.lambda_api_javaagent.id
  resource_id = aws_api_gateway_resource.lambda_api_proxy_javaagent.id
  http_method = "ANY"
  authorization = "NONE"
}

resource "aws_api_gateway_integration" "lambda_api_javaagent" {
  rest_api_id = aws_api_gateway_rest_api.lambda_api_javaagent.id
  resource_id = aws_api_gateway_method.lambda_api_proxy_javaagent.resource_id
  http_method = aws_api_gateway_method.lambda_api_proxy_javaagent.http_method

  integration_http_method = "POST"
  type = "AWS_PROXY"
  uri = aws_lambda_alias.provisioned.invoke_arn
}

resource "aws_api_gateway_method" "lambda_api_proxy_root_javaagent" {
  rest_api_id = aws_api_gateway_rest_api.lambda_api_javaagent.id
  resource_id = aws_api_gateway_rest_api.lambda_api_javaagent.root_resource_id
  http_method = "ANY"
  authorization = "NONE"
}

resource "aws_api_gateway_integration" "lambda_api_root_javaagent" {
  rest_api_id = aws_api_gateway_rest_api.lambda_api_javaagent.id
  resource_id = aws_api_gateway_method.lambda_api_proxy_root_javaagent.resource_id
  http_method = aws_api_gateway_method.lambda_api_proxy_root_javaagent.http_method

  integration_http_method = "POST"
  type = "AWS_PROXY"
  uri = aws_lambda_alias.provisioned.invoke_arn
}

resource "aws_api_gateway_deployment" "lambda_api_javaagent" {
  depends_on = [
    aws_api_gateway_integration.lambda_api_javaagent,
    aws_api_gateway_integration.lambda_api_root_javaagent,
  ]

  rest_api_id = aws_api_gateway_rest_api.lambda_api_javaagent.id
}

resource "aws_api_gateway_stage" "test_javaagent" {
  stage_name = "default"
  rest_api_id = aws_api_gateway_rest_api.lambda_api_javaagent.id
  deployment_id = aws_api_gateway_deployment.lambda_api_javaagent.id
}

resource "aws_lambda_permission" "lambda_api_allow_gateway_javaagent" {
  action = "lambda:InvokeFunction"
  function_name = aws_lambda_alias.provisioned.function_name
  qualifier = aws_lambda_alias.provisioned.name
  principal = "apigateway.amazonaws.com"
  source_arn = "${aws_api_gateway_rest_api.lambda_api_javaagent.execution_arn}/*/*"
}

output "lambda_api_gateway_javaagent_url" {
  value = aws_api_gateway_stage.test_javaagent.invoke_url
}
