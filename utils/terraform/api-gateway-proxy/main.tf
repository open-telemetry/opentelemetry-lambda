resource "aws_api_gateway_rest_api" "lambda_api_proxy" {
  name = var.name
}

resource "aws_api_gateway_resource" "lambda_api_proxy" {
  rest_api_id = aws_api_gateway_rest_api.lambda_api_proxy.id
  parent_id   = aws_api_gateway_rest_api.lambda_api_proxy.root_resource_id
  path_part   = "{proxy+}"
}

resource "aws_api_gateway_method" "lambda_api_proxy" {
  rest_api_id   = aws_api_gateway_rest_api.lambda_api_proxy.id
  resource_id   = aws_api_gateway_resource.lambda_api_proxy.id
  http_method   = "ANY"
  authorization = "NONE"
}

resource "aws_api_gateway_integration" "lambda_api" {
  rest_api_id = aws_api_gateway_rest_api.lambda_api_proxy.id
  resource_id = aws_api_gateway_method.lambda_api_proxy.resource_id
  http_method = aws_api_gateway_method.lambda_api_proxy.http_method

  integration_http_method = "POST"
  type                    = "AWS_PROXY"
  uri                     = var.function_invoke_arn
}

resource "aws_api_gateway_method" "lambda_api_proxy_root_nodejs" {
  rest_api_id   = aws_api_gateway_rest_api.lambda_api_proxy.id
  resource_id   = aws_api_gateway_rest_api.lambda_api_proxy.root_resource_id
  http_method   = "ANY"
  authorization = "NONE"
}

resource "aws_api_gateway_integration" "lambda_api_root_nodejs" {
  rest_api_id = aws_api_gateway_rest_api.lambda_api_proxy.id
  resource_id = aws_api_gateway_method.lambda_api_proxy_root_nodejs.resource_id
  http_method = aws_api_gateway_method.lambda_api_proxy_root_nodejs.http_method

  integration_http_method = "POST"
  type                    = "AWS_PROXY"
  uri                     = var.function_invoke_arn
}

resource "aws_api_gateway_deployment" "lambda_api_proxy" {
  depends_on = [
    aws_api_gateway_integration.lambda_api,
    aws_api_gateway_integration.lambda_api_root_nodejs,
  ]

  rest_api_id = aws_api_gateway_rest_api.lambda_api_proxy.id
}

resource "aws_api_gateway_stage" "test" {
  stage_name           = "default"
  rest_api_id          = aws_api_gateway_rest_api.lambda_api_proxy.id
  deployment_id        = aws_api_gateway_deployment.lambda_api_proxy.id
  xray_tracing_enabled = var.enable_xray_tracing
}

resource "aws_lambda_permission" "lambda_api_allow_gateway_nodejs" {
  action        = "lambda:InvokeFunction"
  function_name = var.function_name
  qualifier     = var.function_qualifier
  principal     = "apigateway.amazonaws.com"
  source_arn    = "${aws_api_gateway_rest_api.lambda_api_proxy.execution_arn}/*/*"
}
