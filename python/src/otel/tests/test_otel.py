import sys
import os

awsLambdaEnvDict = {
    "AWS_EXECUTION_ENV": "AWS_Lambda_python3.8",
    "AWS_LAMBDA_FUNCTION_MEMORY_SIZE": "128",
    "AWS_LAMBDA_FUNCTION_NAME": "python-lambda-function-YI0MC6JQ4BMR",
    "AWS_LAMBDA_FUNCTION_VERSION": "2",
    "AWS_LAMBDA_LOG_GROUP_NAME": "/aws/lambda/python-lambda-function-YI0MC6JQ4BMR",
    "AWS_LAMBDA_LOG_STREAM_NAME": "2020/10/06/[$LATEST]33f5c2beeb3a46dda4e9712885809a22",
    "AWS_LAMBDA_RUNTIME_API": "127.0.0.1:9001",
    "AWS_REGION": "us-east-1",
    "AWS_XRAY_CONTEXT_MISSING": "LOG_ERROR",
    "AWS_XRAY_DAEMON_ADDRESS": "localhost:2000",
    "LAMBDA_RUNTIME_DIR": "/var/runtime",
    "LAMBDA_TASK_ROOT": "/var/task",
    "LANG": "en_US.UTF-8",
    "LD_LIBRARY_PATH": "/var/lang/lib:/lib64:/usr/lib64:/var/runtime:/var/runtime/lib:/var/task:/var/task/lib:/opt/lib",
    "PATH": "/var/lang/bin:/usr/local/bin:/usr/bin/:/bin:/opt/bin",
    "PWD": "/var/task",
    "PYTHONPATH": "/var/runtime",
    "SHLVL": "0",
    "TZ": ":UTC",
    "_AWS_XRAY_DAEMON_ADDRESS": "169.254.79.2",
    "_AWS_XRAY_DAEMON_PORT": "2000",
    "_HANDLER": "lambda_function.lambda_handler",
    "_X_AMZN_TRACE_ID": "Root=1-5fb73311-05e8bb83207fa31d4d9cdb4c;Parent=3328b8445a6dbad2;Sampled=1",
}
for k, v in awsLambdaEnvDict.items():
    os.environ[k] = v

os.environ["ORIG_HANDLER"] = "mock_lambda.handler"
sys.path.append("otel/otel_sdk")

from otel_wrapper import lambda_handler
from opentelemetry import trace
from opentelemetry.trace import SpanKind
from opentelemetry.sdk.trace.export import (
    SimpleExportSpanProcessor,
    ConsoleSpanExporter,
)
from opentelemetry.sdk.trace.export.in_memory_span_exporter import (
    InMemorySpanExporter,
)
from opentelemetry.sdk.trace import TracerProvider
from opentelemetry.resource import AwsLambdaResourceDetector
from opentelemetry.sdk.resources import Resource

class MockLambdaContext:
    pass

lambdaContext = MockLambdaContext()
lambdaContext.invoked_function_arn = "arn://mock-lambda-function-arn"
lambdaContext.aws_request_id = "mock_aws_request_id"

# TODO: does not work, need upstream fix
resource = Resource.create().merge(AwsLambdaResourceDetector().detect())
trace.set_tracer_provider(
    TracerProvider(
        resource=resource,
    )
)
trace.get_tracer_provider().add_span_processor(
    SimpleExportSpanProcessor(ConsoleSpanExporter()),
)


in_memory_exporter = InMemorySpanExporter()
trace.get_tracer_provider().add_span_processor(
    SimpleExportSpanProcessor(in_memory_exporter)
)


def test_lambda_instrument():
    in_memory_exporter.clear()
    lambda_handler("mock", lambdaContext)
    spans = in_memory_exporter.get_finished_spans()
    assert len(spans) == 1

    span = spans[0]
    assert span.name == "mock_lambda.handler"

    parent_context = span.parent
    assert parent_context.trace_id == 0x5FB7331105E8BB83207FA31D4D9CDB4C
    assert parent_context.span_id == 0x3328B8445A6DBAD2
    assert parent_context.is_remote is True

    assert span.context.trace_id == 0x5FB7331105E8BB83207FA31D4D9CDB4C

    assert span.kind == SpanKind.CONSUMER

    # TODO: waiting OTel Python supports env variable for resource detector
    resource_atts = span.resource.attributes
    # assert resource_atts["faas.name"] == "python-lambda-function-YI0MC6JQ4BMR"
    # assert resource_atts["cloud.region"] == "us-east-1"
    # assert resource_atts["cloud.provider"] == "aws"
    # assert resource_atts["faas.version"] == "2"
    assert resource_atts["telemetry.sdk.language"] == "python"
    assert resource_atts["telemetry.sdk.name"] == "opentelemetry"

    attributs = span.attributes
    assert attributs["faas.execution"] == "mock_aws_request_id"
    assert attributs["faas.id"] == "arn://mock-lambda-function-arn"
