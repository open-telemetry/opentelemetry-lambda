variable "name" {
  type        = string
  description = "Name of created function and API Gateway"
  default     = "hello-python"
}

variable "collector_layer_arn" {
  type        = string
  description = "ARN for the Lambda layer containing the OpenTelemetry collector extension"
  // TODO(anuraaga): Add default when a public layer is published.
}

variable "python_wrapper_layer_arn" {
  type        = string
  description = "ARN for the Lambda layer containing the OpenTelemetry Python Wrapper"
  // TODO(anuraaga): Add default when a public layer is published.
}
