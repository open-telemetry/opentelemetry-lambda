resource "aws_lambda_layer_version" "opentelemetry_java_wrapper" {
  layer_name = "opentelemetry-java-wrapper"
  filename = "../../../../layer-wrapper/build/distributions/opentelemetry-java-wrapper.zip"
  compatible_runtimes = ["java8", "java8.al2", "java11"]
  license_info = "Apache-2.0"
  source_code_hash = filebase64sha256("../../../../layer-wrapper/build/distributions/opentelemetry-java-wrapper.zip")
}

module "hello-okhttp-java-wrapper" {
  source = "terraform-aws-modules/lambda/aws"

  function_name = "hello-okhttp-java-wrapper"
  handler       = "io.opentelemetry.lambda.sampleapps.okhttp.OkHttpRequestHandler::handleRequest"
  runtime = "java11"

  create_package         = false
  local_existing_package = "../../build/libs/okhttp-all.jar"

  memory_size = 384
  timeout = 20

  layers = [
    aws_lambda_layer_version.opentelemetry_java_wrapper.arn
  ]

  environment_variables = {
    AWS_LAMBDA_EXEC_WRAPPER = "/opt/otel-proxy-handler"
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

resource "aws_api_gateway_rest_api" "lambda_api_java_wrapper" {
  name = "hello-lambda-okhttp-java-wrapper"
}

resource "aws_api_gateway_resource" "lambda_api_proxy_java_wrapper" {
  rest_api_id = aws_api_gateway_rest_api.lambda_api_java_wrapper.id
  parent_id = aws_api_gateway_rest_api.lambda_api_java_wrapper.root_resource_id
  path_part = "{proxy+}"
}

resource "aws_api_gateway_method" "lambda_api_proxy_java_wrapper" {
  rest_api_id = aws_api_gateway_rest_api.lambda_api_java_wrapper.id
  resource_id = aws_api_gateway_resource.lambda_api_proxy_java_wrapper.id
  http_method = "ANY"
  authorization = "NONE"
}

resource "aws_api_gateway_integration" "lambda_api_java_wrapper" {
  rest_api_id = aws_api_gateway_rest_api.lambda_api_java_wrapper.id
  resource_id = aws_api_gateway_method.lambda_api_proxy_java_wrapper.resource_id
  http_method = aws_api_gateway_method.lambda_api_proxy_java_wrapper.http_method

  integration_http_method = "POST"
  type = "AWS_PROXY"
  uri = module.hello-okhttp-java-wrapper.this_lambda_function_invoke_arn
}

resource "aws_api_gateway_method" "lambda_api_proxy_root_java_wrapper" {
  rest_api_id = aws_api_gateway_rest_api.lambda_api_java_wrapper.id
  resource_id = aws_api_gateway_rest_api.lambda_api_java_wrapper.root_resource_id
  http_method = "ANY"
  authorization = "NONE"
}

resource "aws_api_gateway_integration" "lambda_api_root_java_wrapper" {
  rest_api_id = aws_api_gateway_rest_api.lambda_api_java_wrapper.id
  resource_id = aws_api_gateway_method.lambda_api_proxy_root_java_wrapper.resource_id
  http_method = aws_api_gateway_method.lambda_api_proxy_root_java_wrapper.http_method

  integration_http_method = "POST"
  type = "AWS_PROXY"
  uri = module.hello-okhttp-java-wrapper.this_lambda_function_invoke_arn
}

resource "aws_api_gateway_deployment" "lambda_api_java_wrapper" {
  depends_on = [
    aws_api_gateway_integration.lambda_api_java_wrapper,
    aws_api_gateway_integration.lambda_api_root_java_wrapper,
  ]

  rest_api_id = aws_api_gateway_rest_api.lambda_api_java_wrapper.id
}

resource "aws_api_gateway_stage" "test_java_wrapper" {
  stage_name = "default"
  rest_api_id = aws_api_gateway_rest_api.lambda_api_java_wrapper.id
  deployment_id = aws_api_gateway_deployment.lambda_api_java_wrapper.id
}

resource "aws_lambda_permission" "lambda_api_allow_gateway_java_wrapper" {
  action = "lambda:InvokeFunction"
  function_name = module.hello-okhttp-java-wrapper.this_lambda_function_name
  principal = "apigateway.amazonaws.com"
  source_arn = "${aws_api_gateway_rest_api.lambda_api_java_wrapper.execution_arn}/*/*"
}

output "lambda_api_gateway_java_wrapper_url" {
  value = aws_api_gateway_stage.test_java_wrapper.invoke_url
}
