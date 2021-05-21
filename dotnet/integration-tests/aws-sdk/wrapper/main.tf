resource "aws_lambda_layer_version" "collector_layer" {
  count               = var.enable_collector_layer ? 1 : 0
  layer_name          = var.collector_layer_name
  filename            = "${path.module}/../../../../collector/build/collector-extension.zip"
  compatible_runtimes = ["dotnetcore2.1", "dotnetcore3.1"]
  license_info        = "Apache-2.0"
  source_code_hash    = filebase64sha256("${path.module}/../../../../collector/build/collector-extension.zip")
}

module "hello-lambda-function" {
  source              = "../../../sample-apps/aws-sdk/deploy/wrapper"
  name                = var.function_name
  collector_layer_arn = var.enable_collector_layer ? aws_lambda_layer_version.collector_layer[0].arn : null
  tracing_mode        = var.tracing_mode
}

resource "aws_iam_role_policy_attachment" "hello-lambda-cloudwatch-insights" {
  role       = module.hello-lambda-function.function_role_name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
}
