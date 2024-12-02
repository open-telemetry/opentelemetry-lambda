module "hello-lambda-function" {
  source  = "terraform-aws-modules/lambda/aws"
  version = ">= 2.24.0"

  architectures = compact([var.architecture])
  function_name = var.name
  handler       = "io.opentelemetry.lambda.sampleapps.sqs.SqsRequestHandler::handleRequest"
  runtime       = var.runtime

  create_package         = false
  local_existing_package = "${path.module}/../../build/libs/sqs-all.jar"

  memory_size = 512
  timeout     = 120
  publish     = true

  layers = compact([
    var.collector_layer_arn,
    var.sdk_layer_arn,
    var.collector_config_layer_arn,
  ])

  environment_variables = (var.collector_config_layer_arn == null ?
    {
      AWS_LAMBDA_EXEC_WRAPPER = "/opt/otel-handler",
      OTEL_METRICS_EXPORTER   = "otlp",
    } :
    {
      AWS_LAMBDA_EXEC_WRAPPER            = "/opt/otel-handler",
      OTEL_METRICS_EXPORTER              = "otlp",
      OPENTELEMETRY_COLLECTOR_CONFIG_URI = "/opt/config.yaml"
  })

  tracing_mode = var.tracing_mode

  # UNCOMMENT BELOW TO TEST WITH YOUR SQS QUEUE
  # policy_statements = {
  #   sqs_read = {
  #     effect    = "Allow",
  #     actions   = ["sqs:ReceiveMessage", "sqs:DeleteMessage", "sqs:GetQueueAttributes"]
  #     resources = [var.sqs_queue_arn]
  #   }
  # }

  # event_source_mapping = {
  #   sqs_queue = {
  #     event_source_arn = var.sqs_queue_arn
  #   }
  # }

  # allowed_triggers = {
  #   sqs_queue = {
  #     principal  = "sqs.amazonaws.com"
  #     source_arn = "${var.sqs_queue_arn}"
  #   }
  # }
}
