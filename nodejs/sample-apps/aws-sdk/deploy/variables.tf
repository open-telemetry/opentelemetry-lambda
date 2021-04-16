variable "name" {
  type        = string
  description = "Name of created function and API Gateway"
  default     = "hello-nodejs-awssdk"
}

variable "collector_layer_arn" {
  type        = string
  description = "ARN for the Lambda layer containing the OpenTelemetry collector extension"
  // TODO(anuraaga): Add default when a public layer is published.
}

variable "nodejs_wrapper_layer_arn" {
  type        = string
  description = "ARN for the Lambda layer containing the OpenTelemetry NodeJS Wrapper"
  // TODO(anuraaga): Add default when a public layer is published.
}
