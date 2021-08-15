variable "collector_layer_name" {
  type        = string
  description = "Name of published collector layer"
  default     = "opentelemetry-collector"
}

variable "sdk_layer_name" {
  type        = string
  description = "Name of published SDK layer"
  default     = "opentelemetry-java-agent"
}

variable "collector_config_layer_arn" {
  type        = string
  description = "Name of published SDK layer"
}

variable "function_name" {
  type        = string
  description = "Name of sample app function / API gateway"
  default     = "hello-java-awssdk-agent"
}

variable "tracing_mode" {
  type        = string
  description = "Lambda function tracing mode"
  default     = "PassThrough"
}

variable "enable_collector_layer" {
  type        = bool
  description = "Enables building and usage of a layer for the collector. If false, it means either the SDK layer includes the collector or it is not used."
  default     = true
}
