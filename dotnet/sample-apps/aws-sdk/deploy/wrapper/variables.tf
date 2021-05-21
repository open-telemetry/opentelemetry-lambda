variable "name" {
  type        = string
  description = "Name of created function and API Gateway"
  default     = "hello-dotnet-awssdk-wrapper"
}

variable "collector_layer_arn" {
  type        = string
  description = "ARN for the Lambda layer containing the OpenTelemetry collector extension"
}

variable "tracing_mode" {
  type        = string
  description = "Lambda function tracing mode"
  default     = "PassThrough"
}
