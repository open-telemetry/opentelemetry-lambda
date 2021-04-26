variable "name" {
  type        = string
  description = "Name of API gateway to create"
}

variable "function_name" {
  type        = string
  description = "Name of function to proxy to"
}

variable "function_qualifier" {
  type        = string
  default     = null
  description = "Qualifier of function to proxy to"
}

variable "function_invoke_arn" {
  type        = string
  description = "Invoke ARN of function to proxy to"
}

variable "enable_xray_tracing" {
  type        = bool
  description = "Whether to enable xray tracing of the API gateway"
  default     = false
}
