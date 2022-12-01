resource "aws_lambda_layer_version" "sdk_layer" {
  layer_name          = var.sdk_layer_name
  filename            = "${path.module}/../../../src/build/layer.zip"
  compatible_runtimes = ["python3.8", "python3.9"]
  license_info        = "Apache-2.0"
  source_code_hash    = filebase64sha256("${path.module}/../../../src/build/layer.zip")
}

resource "aws_lambda_layer_version" "collector_layer" {
  count               = var.enable_collector_layer ? 1 : 0
  layer_name          = var.collector_layer_name
  filename            = "${path.module}/../../../../collector/build/collector-extension.zip"
  compatible_runtimes = ["nodejs14.x", "nodejs16.x", "nodejs18.x"]
  license_info        = "Apache-2.0"
  source_code_hash    = filebase64sha256("${path.module}/../../../../collector/build/collector-extension.zip")
}

module "hello-lambda-function" {
  source              = "../../../sample-apps/aws-sdk/deploy/wrapper"
  name                = var.function_name
  architecture        = var.architecture
  collector_layer_arn = var.enable_collector_layer ? aws_lambda_layer_version.collector_layer[0].arn : null
  sdk_layer_arn       = aws_lambda_layer_version.sdk_layer.arn
  tracing_mode        = var.tracing_mode
  runtime             = var.runtime
}

resource "aws_iam_role_policy_attachment" "hello-lambda-cloudwatch-insights" {
  role       = module.hello-lambda-function.function_role_name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
}
