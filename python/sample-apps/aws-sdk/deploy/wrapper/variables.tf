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

variable "sdk_layer_arn" {
  type        = string
  description = "ARN for the Lambda layer containing the OpenTelemetry Python Wrapper"
  // TODO(anuraaga): Add default when a public layer is published.
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

variable "runtime" {
  type        = string
  description = "Python runtime version used for sample Lambda Function"
  default     = "python3.9"
}
