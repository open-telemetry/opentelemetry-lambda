resource "aws_lambda_layer_version" "sdk_layer" {
  layer_name          = var.sdk_layer_name
  filename            = "../../packages/layer/build/layer.zip"
  compatible_runtimes = ["nodejs10.x", "nodejs12.x", "nodejs14.x"]
  license_info        = "Apache-2.0"
  source_code_hash    = filebase64sha256("../../packages/layer/build/layer.zip")
}

resource "aws_lambda_layer_version" "collector_layer" {
  layer_name          = var.collector_layer_name
  filename            = "../../../collector/build/collector-extension.zip"
  compatible_runtimes = ["nodejs10.x", "nodejs12.x", "nodejs14.x"]
  license_info        = "Apache-2.0"
  source_code_hash    = filebase64sha256("../../../collector/build/collector-extension.zip")
}

module "hello-awssdk-function" {
  source                   = "../../sample-apps/aws-sdk/deploy"
  name                     = var.function_name
  collector_layer_arn      = aws_lambda_layer_version.sdk_layer.arn
  nodejs_wrapper_layer_arn = aws_lambda_layer_version.collector_layer.arn
  enable_lambda_insights   = var.enable_lambda_insights
}
