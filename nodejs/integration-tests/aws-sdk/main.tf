resource "aws_lambda_layer_version" "sdk_layer" {
  layer_name          = var.sdk_layer_name
  filename            = "${path.module}/../../packages/layer/build/layer.zip"
  compatible_runtimes = ["nodejs10.x", "nodejs12.x", "nodejs14.x"]
  license_info        = "Apache-2.0"
  source_code_hash    = filebase64sha256("${path.module}/../../packages/layer/build/layer.zip")
}

resource "aws_lambda_layer_version" "collector_layer" {
  layer_name          = var.collector_layer_name
  filename            = "${path.module}/../../../collector/build/collector-extension.zip"
  compatible_runtimes = ["nodejs10.x", "nodejs12.x", "nodejs14.x"]
  license_info        = "Apache-2.0"
  source_code_hash    = filebase64sha256("${path.module}/../../../collector/build/collector-extension.zip")
}

module "hello-lambda-function" {
  source              = "../../sample-apps/aws-sdk/deploy"
  name                = var.function_name
  collector_layer_arn = aws_lambda_layer_version.sdk_layer.arn
  sdk_layer_arn       = aws_lambda_layer_version.collector_layer.arn
  tracing_mode        = var.tracing_mode
}
