variable "name" {
  type        = string
  description = "Name of created function and API Gateway"
  default     = "hello-java-awssdk-wrapper"
}

variable "collector_layer_arn" {
  type        = string
  description = "ARN for the Lambda layer containing the OpenTelemetry collector extension"
  // TODO(anuraaga): Add default when a public layer is published.
}

variable "sdk_layer_arn" {
  type        = string
  description = "ARN for the Lambda layer containing the OpenTelemetry Java Wrapper"
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

# UNCOMMENT BELOW TO TEST WITH YOUR SQS QUEUE
# variable "sqs_queue_arn" {
#   type        = string
#   description = "ARN for the SQS queue to use an event source for the Lambda"
# }

variable "runtime" {
  type        = string
  description = "java runtime used for sample Lambda Function"
  default     = "java11"
}
