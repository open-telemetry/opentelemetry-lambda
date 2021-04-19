resource "aws_lambda_layer_version" "opentelemetry_python_wrapper" {
  layer_name          = var.sdk_layer_name
  filename            = "../../src/build/layer.zip"
  compatible_runtimes = ["python3.8"]
  license_info        = "Apache-2.0"
  source_code_hash    = filebase64sha256("../../src/build/layer.zip")
}

resource "aws_lambda_layer_version" "opentelemetry_collector" {
  layer_name          = var.collector_layer_name
  filename            = "../../../collector/build/collector-extension.zip"
  compatible_runtimes = ["nodejs10.x", "nodejs12.x", "nodejs14.x"]
  license_info        = "Apache-2.0"
  source_code_hash    = filebase64sha256("../../../collector/build/collector-extension.zip")
}

module "function" {
  source                   = "../../sample-apps/deploy"
  name                     = var.function_name
  collector_layer_arn      = aws_lambda_layer_version.opentelemetry_collector.arn
  python_wrapper_layer_arn = aws_lambda_layer_version.opentelemetry_python_wrapper.arn
}
