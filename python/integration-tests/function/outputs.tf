output "api-gateway-url" {
  value = module.hello-lambda-function.api-gateway-url
}

output "function_role_name" {
  value = module.hello-lambda-function.function_role_name
}
