resource "aws_lambda_layer_version" "opentelemetry_nodejs_wrapper" {
  layer_name          = "opentelemetry-nodejs-wrapper"
  filename            = "../../packages/layer/build/layer.zip"
  compatible_runtimes = ["nodejs10.x", "nodejs12.x", "nodejs14.x"]
  license_info        = "Apache-2.0"
  source_code_hash    = filebase64sha256("../../packages/layer/build/layer.zip")
}

resource "aws_lambda_layer_version" "opentelemetry_collector" {
  layer_name          = "opentelemetry-collector"
  filename            = "../../../collector/build/collector-extension.zip"
  compatible_runtimes = ["nodejs10.x", "nodejs12.x", "nodejs14.x"]
  license_info        = "Apache-2.0"
  source_code_hash    = filebase64sha256("../../../collector/build/collector-extension.zip")
}

module "hello-awssdk-function" {
  source                   = "../../sample-apps/aws-sdk/deploy"
  collector_layer_arn      = aws_lambda_layer_version.opentelemetry_collector.arn
  nodejs_wrapper_layer_arn = aws_lambda_layer_version.opentelemetry_nodejs_wrapper.arn
}

output "api-gateway-url" {
  value = module.hello-awssdk-function.api-gateway-url
}
