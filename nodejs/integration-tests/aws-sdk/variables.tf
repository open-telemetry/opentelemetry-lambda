variable "collector_layer_name" {
  type        = string
  description = "Name of published collector layer"
  default     = "opentelemetry-collector"
}

variable "sdk_layer_name" {
  type        = string
  description = "Name of published SDK layer"
  default     = "opentelemetry-nodejs-wrapper"
}

variable "function_name" {
  type        = string
  description = "Name of sample app function / API gateway"
  default     = "hello-nodejs-awssdk"
}

variable "enable_lambda_insights" {
  type        = bool
  description = "Whether to enable Lambda Insights. For now, only supports enabling on us-east-1"
  default     = false
}
