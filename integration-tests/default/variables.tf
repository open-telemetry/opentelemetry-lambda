variable "collector_layer_name" {
  type        = string
  description = "Name of published collector layer"
  default     = "opentelemetry-collector"
}

variable "collector_layer_zip_path" {
  type        = string
  description = "The relative path to the Collector Layer .zip file from the terraform directory."
  default     = "invalid"
}

variable "collector_layer_compatible_runtimes" {
  type        = list(string)
  description = "The compatible runtimes for the Collector Lambda Layer."
  default     = ["invalid"]
}

variable "collector_config_layer_arn" {
  type        = string
  description = "ARN for the Lambda layer containing the OpenTelemetry collector configuration file"
}

variable "sdk_layer_name" {
  type        = string
  description = "Name of published SDK layer"
  default     = "opentelemetry-unknown-wrapper"
}

variable "sdk_layer_zip_path" {
  type        = string
  description = "The relative path to the SDK Layer .zip file from the terraform directory."
  default     = "invalid"
}

variable "sdk_layer_compatible_runtimes" {
  type        = list(string)
  description = "The compatible runtimes for the SDK Lambda Layer."
  default     = ["invalid"]
}

variable "function_name" {
  type        = string
  description = "Name of sample app function / API gateway"
  default     = "unknown"
}

variable "function_terraform_source_path" {
  type        = string
  description = "The path to the terraform directory that creates a sample Lambda function"
  default     = "invalid"
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
