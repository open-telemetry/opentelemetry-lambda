variable "name" {
  type        = string
  description = "Name of created function and API Gateway"
  default     = "hello-go-awssdk-wrapper"
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

variable "architecture" {
  type        = string
  description = "Lambda function architecture, valid values are arm64 or x86_64"
  default     = "x86_64"
}
