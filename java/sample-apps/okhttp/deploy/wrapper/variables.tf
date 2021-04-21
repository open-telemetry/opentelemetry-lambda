variable "name" {
  type        = string
  description = "Name of created function and API Gateway"
  default     = "hello-java-okhttp-wrapper"
}

variable "collector_layer_arn" {
  type        = string
  description = "ARN for the Lambda layer containing the OpenTelemetry collector extension"
  // TODO(anuraaga): Add default when a public layer is published.
}

variable "java_wrapper_layer_arn" {
  type        = string
  description = "ARN for the Lambda layer containing the OpenTelemetry Java Wrapper"
  // TODO(anuraaga): Add default when a public layer is published.
}

variable "enable_lambda_insights" {
  type        = bool
  description = "Whether to enable Lambda Insights. For now, only supports enabling on us-east-1"
  default     = false
}
