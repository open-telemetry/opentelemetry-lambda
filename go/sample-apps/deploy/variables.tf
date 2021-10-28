variable "name" {
  type        = string
  description = "Name of created function and API Gateway"
  default     = "hello-go-awssdk-wrapper"
}

variable "collector_layer_arn" {
  type        = string
  description = "ARN for the Lambda layer containing the OpenTelemetry collector extension"
  default     = "arn:aws:lambda:us-east-1:901920570463:layer:aws-otel-collector-ver-0-29-1:1"
}

variable "tracing_mode" {
  type        = string
  description = "Lambda function tracing mode"
  default     = "PassThrough"
}
