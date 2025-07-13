variable "name" {
  type        = string
  description = "Name of created function and API Gateway"
  default     = "hello-nodejs-awssdk"
}

variable "account_id" {
  type        = string
  description = "AWS account ID where the Lambda layers are published"
  default     = "184161586896"
}

variable "tracing_mode" {
  type        = string
  description = "Lambda function tracing mode"
  default     = "PassThrough"
}

variable "architecture" {
  type        = string
  description = "Lambda function architecture, valid values are arm64 or x86_64"
  default     = "arm64"
}

variable "runtime" {
  type        = string
  description = "NodeJS runtime version used for sample Lambda Function"
  default     = "nodejs22.x"
}

variable "collector_layer_version" {
  type        = string
  description = "Collector layer version, see latest releases here: https://github.com/open-telemetry/opentelemetry-lambda/releases"
  default     = "0_15_0"
}

variable "nodejs_layer_version" {
  type        = string
  description = "Node.js layer version, see latest releases here: https://github.com/open-telemetry/opentelemetry-lambda/releases"
  default     = "0_14_0"
}