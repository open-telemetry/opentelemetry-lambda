# Copyright The OpenTelemetry Authors
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
import os
import sys
from importlib import import_module
from unittest import mock

from opentelemetry.instrumentation.aws_lambda import AwsLambdaInstrumentor
from opentelemetry.sdk.extension.aws.trace.propagation.aws_xray_format import (
    TRACE_ID_FIRST_PART_LENGTH,
)
from opentelemetry.semconv.resource import ResourceAttributes
from opentelemetry.semconv.trace import SpanAttributes
from opentelemetry.test.test_base import TestBase
from opentelemetry.trace import SpanKind

AWS_LAMBDA_EXEC_WRAPPER = "AWS_LAMBDA_EXEC_WRAPPER"
_HANDLER = "_HANDLER"
INSTRUMENTATION_SRC_DIR = os.path.join(
    *(os.path.dirname(__file__), "..", "otel_sdk")
)


class MockLambdaContext:
    def __init__(self, aws_request_id, invoked_function_arn):
        self.invoked_function_arn = invoked_function_arn
        self.aws_request_id = aws_request_id


MOCK_LAMBDA_CONTEXT = MockLambdaContext(
    aws_request_id="mock_aws_request_id",
    invoked_function_arn="arn://mock-lambda-function-arn",
)
MOCK_TRACE_ID = 0x5FB7331105E8BB83207FA31D4D9CDB4C
MOCK_TRACE_ID_HEX_STR = f"{MOCK_TRACE_ID:032x}"
MOCK_PARENT_SPAN_ID = 0x3328B8445A6DBAD2
MOCK_PARENT_SPAN_ID_STR = f"{MOCK_PARENT_SPAN_ID:32x}"
MOCK_LAMBDA_TRACE_CONTEXT_SAMPLED = f"Root=1-{MOCK_TRACE_ID_HEX_STR[:TRACE_ID_FIRST_PART_LENGTH]}-{MOCK_TRACE_ID_HEX_STR[TRACE_ID_FIRST_PART_LENGTH:]};Parent={MOCK_PARENT_SPAN_ID_STR};Sampled=1"
MOCK_LAMBDA_TRACE_CONTEXT_NOT_SAMPLED = f"Root=1-{MOCK_TRACE_ID_HEX_STR[:TRACE_ID_FIRST_PART_LENGTH]}-{MOCK_TRACE_ID_HEX_STR[TRACE_ID_FIRST_PART_LENGTH:]};Parent={MOCK_PARENT_SPAN_ID};Sampled=0"


def mock_aws_lambda_exec_wrapper():
    """Mocks automatically instrumenting user Lambda function by pointing
    `AWS_LAMBDA_EXEC_WRAPPER` to the `otel-instrument` script.

    TODO: It would be better if `moto`'s `mock_lambda` supported setting
    AWS_LAMBDA_EXEC_WRAPPER so we could make the call to Lambda instead.

    See more:
    https://aws-otel.github.io/docs/getting-started/lambda/lambda-python
    """
    exec(open(os.path.join(INSTRUMENTATION_SRC_DIR, "otel-instrument")).read())


def mock_execute_lambda():
    if os.environ[AWS_LAMBDA_EXEC_WRAPPER]:
        globals()[os.environ[AWS_LAMBDA_EXEC_WRAPPER]]()

    module_name, handler_name = os.environ[_HANDLER].split(".")
    handler_module = import_module(".".join(module_name.split("/")))
    getattr(handler_module, handler_name)("mock_event", MOCK_LAMBDA_CONTEXT)


class TestAwsLambdaInstrumentor(TestBase):
    """AWS Lambda Instrumentation Testsuite"""

    @classmethod
    def setUpClass(cls):
        super().setUpClass()
        sys.path.append(INSTRUMENTATION_SRC_DIR)

    def setUp(self):
        super().setUp()
        self.common_env_patch = mock.patch.dict(
            "os.environ",
            {
                AWS_LAMBDA_EXEC_WRAPPER: "mock_aws_lambda_exec_wrapper",
                "AWS_LAMBDA_FUNCTION_NAME": "test-python-lambda-function",
                "AWS_LAMBDA_FUNCTION_VERSION": "2",
                "AWS_REGION": "us-east-1",
                _HANDLER: "mock_user_lambda.handler",
            },
        )
        self.common_env_patch.start()

    def tearDown(self):
        super().tearDown()
        self.common_env_patch.stop()
        AwsLambdaInstrumentor().uninstrument()

    @classmethod
    def tearDownClass(cls):
        super().tearDownClass()
        sys.path.remove(INSTRUMENTATION_SRC_DIR)

    def test_active_tracing(self):
        test_env_patch = mock.patch.dict(
            "os.environ",
            {
                **os.environ,
                "_X_AMZN_TRACE_ID": MOCK_LAMBDA_TRACE_CONTEXT_SAMPLED,
            },
        )
        test_env_patch.start()

        mock_execute_lambda()

        spans = self.memory_exporter.get_finished_spans()

        assert spans

        self.assertEqual(len(spans), 1)
        span = spans[0]
        self.assertEqual(span.name, os.environ["ORIG_HANDLER"])
        self.assertEqual(span.get_span_context().trace_id, MOCK_TRACE_ID)
        self.assertEqual(span.kind, SpanKind.SERVER)
        self.assertSpanHasAttributes(
            span,
            {
                ResourceAttributes.FAAS_ID: MOCK_LAMBDA_CONTEXT.invoked_function_arn,
                SpanAttributes.FAAS_EXECUTION: MOCK_LAMBDA_CONTEXT.aws_request_id,
            },
        )

        # TODO: waiting OTel Python supports env variable for resource detector
        # resource_atts = span.resource.attributes
        # assert resource_atts["faas.name"] == "test-python-lambda-function"
        # assert resource_atts["cloud.region"] == "us-east-1"
        # assert resource_atts["cloud.provider"] == "aws"
        # assert resource_atts["faas.version"] == "2"

        parent_context = span.parent
        self.assertEqual(
            parent_context.trace_id, span.get_span_context().trace_id
        )
        self.assertEqual(parent_context.span_id, MOCK_PARENT_SPAN_ID)
        self.assertTrue(parent_context.is_remote)

        test_env_patch.stop()
