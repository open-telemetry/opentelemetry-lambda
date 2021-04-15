resource "aws_lambda_layer_version" "opentelemetry_java_wrapper" {
  layer_name          = "opentelemetry-java-wrapper"
  filename            = "../../../layer-wrapper/build/distributions/opentelemetry-java-wrapper.zip"
  compatible_runtimes = ["java8", "java8.al2", "java11"]
  license_info        = "Apache-2.0"
  source_code_hash    = filebase64sha256("../../../layer-wrapper/build/distributions/opentelemetry-java-wrapper.zip")
}

resource "aws_lambda_layer_version" "opentelemetry_collector" {
  layer_name          = "opentelemetry-collector"
  filename            = "../../../../collector/build/collector-extension.zip"
  compatible_runtimes = ["nodejs10.x", "nodejs12.x", "nodejs14.x"]
  license_info        = "Apache-2.0"
  source_code_hash    = filebase64sha256("../../../../collector/build/collector-extension.zip")
}

module "hello-okhttp-function" {
  source                 = "../../../sample-apps/okhttp/deploy/wrapper"
  collector_layer_arn    = aws_lambda_layer_version.opentelemetry_collector.arn
  java_wrapper_layer_arn = aws_lambda_layer_version.opentelemetry_java_wrapper.arn
}

output "api-gateway-url" {
  value = module.hello-okhttp-function.api-gateway-url
}
