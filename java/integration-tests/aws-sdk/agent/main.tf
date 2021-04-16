resource "aws_lambda_layer_version" "opentelemetry_javaagent" {
  layer_name          = var.sdk_layer_name
  filename            = "../../../layer-javaagent/build/distributions/opentelemetry-javaagent-layer.zip"
  compatible_runtimes = ["java8", "java8.al2", "java11"]
  license_info        = "Apache-2.0"
  source_code_hash    = filebase64sha256("../../../layer-javaagent/build/distributions/opentelemetry-javaagent-layer.zip")
}

resource "aws_lambda_layer_version" "opentelemetry_collector" {
  layer_name          = var.collector_layer_name
  filename            = "../../../../collector/build/collector-extension.zip"
  compatible_runtimes = ["nodejs10.x", "nodejs12.x", "nodejs14.x"]
  license_info        = "Apache-2.0"
  source_code_hash    = filebase64sha256("../../../../collector/build/collector-extension.zip")
}

module "hello-awssdk-function" {
  source               = "../../../sample-apps/aws-sdk/deploy/agent"
  name                 = var.function_name
  collector_layer_arn  = aws_lambda_layer_version.opentelemetry_collector.arn
  java_agent_layer_arn = aws_lambda_layer_version.opentelemetry_javaagent.arn
}
