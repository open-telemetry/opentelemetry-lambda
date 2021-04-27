output "api-gateway-url" {
  value = module.hello-lambda-function.api-gateway-url
}

output "function_role_name" {
  value = module.hello-lambda-function.function_role_name
}

output "collector_layer_arn" {
  value = var.enable_collector_layer ? aws_lambda_layer_version.collector_layer[0].arn : ""
}

output "sdk_layer_arn" {
  value = aws_lambda_layer_version.sdk_layer.arn
}
