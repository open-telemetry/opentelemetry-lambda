variable "name" {
  type        = string
  description = "Name of created function and API Gateway"
  default     = "hello-java-awssdk-agent"
}

variable "collector_layer_arn" {
  type        = string
  description = "ARN for the Lambda layer containing the OpenTelemetry collector extension"
  // TODO(anuraaga): Add default when a public layer is published.
}

variable "java_agent_layer_arn" {
  type        = string
  description = "ARN for the Lambda layer containing the OpenTelemetry Java Agent"
  // TODO(anuraaga): Add default when a public layer is published.
}
