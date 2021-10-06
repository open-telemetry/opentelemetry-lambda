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
from opentelemetry.propagate import get_global_textmap
from opentelemetry.sdk.extension.aws.trace.propagation.aws_xray_format import (
    TRACE_ID_FIRST_PART_LENGTH,
    TRACE_ID_VERSION,
)
from opentelemetry.semconv.resource import ResourceAttributes
from opentelemetry.semconv.trace import SpanAttributes
from opentelemetry.test.test_base import TestBase
from opentelemetry.trace import SpanKind

_HANDLER = "_HANDLER"
_X_AMZN_TRACE_ID = "_X_AMZN_TRACE_ID"
AWS_LAMBDA_EXEC_WRAPPER = "AWS_LAMBDA_EXEC_WRAPPER"
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

MOCK_XRAY_TRACE_ID = 0x5FB7331105E8BB83207FA31D4D9CDB4C
MOCK_XRAY_TRACE_ID_STR = f"{MOCK_XRAY_TRACE_ID:x}"
MOCK_XRAY_PARENT_SPAN_ID = 0x3328B8445A6DBAD2
MOCK_XRAY_TRACE_CONTEXT_COMMON = f"Root={TRACE_ID_VERSION}-{MOCK_XRAY_TRACE_ID_STR[:TRACE_ID_FIRST_PART_LENGTH]}-{MOCK_XRAY_TRACE_ID_STR[TRACE_ID_FIRST_PART_LENGTH:]};Parent={MOCK_XRAY_PARENT_SPAN_ID:x}"
MOCK_XRAY_TRACE_CONTEXT_SAMPLED = f"{MOCK_XRAY_TRACE_CONTEXT_COMMON};Sampled=1"
MOCK_XRAY_TRACE_CONTEXT_NOT_SAMPLED = (
    f"{MOCK_XRAY_TRACE_CONTEXT_COMMON};Sampled=0"
)

# Read more:
# https://www.w3.org/TR/trace-context/#examples-of-http-traceparent-headers
MOCK_W3C_TRACE_ID = 0x5CE0E9A56015FEC5AADFA328AE398115
MOCK_W3C_PARENT_SPAN_ID = 0xAB54A98CEB1F0AD2
MOCK_W3C_TRACE_CONTEXT_SAMPLED = (
    f"00-{MOCK_W3C_TRACE_ID:x}-{MOCK_W3C_PARENT_SPAN_ID:x}-01"
)

MOCK_W3C_TRACE_STATE_KEY = "vendor_specific_key"
MOCK_W3C_TRACE_STATE_VALUE = "test_value"


def mock_aws_lambda_exec_wrapper():
    """Mocks automatically instrumenting user Lambda function by pointing
    `AWS_LAMBDA_EXEC_WRAPPER` to the `otel-instrument` script.

    TODO: It would be better if `moto`'s `mock_lambda` supported setting
    AWS_LAMBDA_EXEC_WRAPPER so we could make the call to Lambda instead.

    See more:
    https://aws-otel.github.io/docs/getting-started/lambda/lambda-python
    """
    # NOTE: AwsLambdaInstrumentor().instrument() is done at this point
    exec(open(os.path.join(INSTRUMENTATION_SRC_DIR, "otel-instrument")).read())


def mock_execute_lambda(event=None):
    """Mocks Lambda importing and then calling the method at the current
    `_HANDLER` environment variable. Like the real Lambda, if
    `AWS_LAMBDA_EXEC_WRAPPER` is defined, if executes that before `_HANDLER`.

    See more:
    https://aws-otel.github.io/docs/getting-started/lambda/lambda-python
    """
    if os.environ[AWS_LAMBDA_EXEC_WRAPPER]:
        globals()[os.environ[AWS_LAMBDA_EXEC_WRAPPER]]()

    module_name, handler_name = os.environ[_HANDLER].split(".")
    handler_module = import_module(".".join(module_name.split("/")))
    getattr(handler_module, handler_name)(event, MOCK_LAMBDA_CONTEXT)


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
                _X_AMZN_TRACE_ID: MOCK_XRAY_TRACE_CONTEXT_SAMPLED,
            },
        )
        test_env_patch.start()

        mock_execute_lambda()

        spans = self.memory_exporter.get_finished_spans()

        assert spans

        self.assertEqual(len(spans), 1)
        span = spans[0]
        self.assertEqual(span.name, os.environ["ORIG_HANDLER"])
        self.assertEqual(span.get_span_context().trace_id, MOCK_XRAY_TRACE_ID)
        self.assertEqual(span.kind, SpanKind.SERVER)
        self.assertSpanHasAttributes(
            span,
            {
                ResourceAttributes.FAAS_ID: MOCK_LAMBDA_CONTEXT.invoked_function_arn,
                SpanAttributes.FAAS_EXECUTION: MOCK_LAMBDA_CONTEXT.aws_request_id,
            },
        )

        # TODO: Waiting on OTel Python support for setting Resource Detectors
        # using environment variables. Auto Instrumentation (used by this Lambda
        # Instrumentation) sets up the global TracerProvider which is the only
        # time Resource Detectors can be configured.
        #
        # resource_atts = span.resource.attributes
        # self.assertEqual(resource_atts[ResourceAttributes.CLOUD_PLATFORM], CloudPlatformValues.AWS_LAMBDA.value)
        # self.assertEqual(resource_atts[ResourceAttributes.CLOUD_PROVIDER], CloudProviderValues.AWS.value)
        # self.assertEqual(resource_atts[ResourceAttributes.CLOUD_REGION], os.environ["AWS_REGION"])
        # self.assertEqual(resource_atts[ResourceAttributes.FAAS_NAME], os.environ["AWS_LAMBDA_FUNCTION_NAME"])
        # self.assertEqual(resource_atts[ResourceAttributes.FAAS_VERSION], os.environ["AWS_LAMBDA_FUNCTION_VERSION"])

        parent_context = span.parent
        self.assertEqual(
            parent_context.trace_id, span.get_span_context().trace_id
        )
        self.assertEqual(parent_context.span_id, MOCK_XRAY_PARENT_SPAN_ID)
        self.assertTrue(parent_context.is_remote)

        test_env_patch.stop()

    def test_parent_context_from_lambda_event(self):
        test_env_patch = mock.patch.dict(
            "os.environ",
            {
                **os.environ,
                # NOT Active Tracing
                _X_AMZN_TRACE_ID: MOCK_XRAY_TRACE_CONTEXT_NOT_SAMPLED,
                # NOT using the X-Ray Propagator
                "OTEL_PROPAGATORS": "tracecontext",
            },
        )
        test_env_patch.start()

        mock_execute_lambda(
            {
                "headers": {
                    "traceparent": MOCK_W3C_TRACE_CONTEXT_SAMPLED,
                    "tracestate": f"{MOCK_W3C_TRACE_STATE_KEY}={MOCK_W3C_TRACE_STATE_VALUE},foo=1,bar=2",
                }
            }
        )

        spans = self.memory_exporter.get_finished_spans()

        assert spans

        self.assertEqual(len(spans), 1)
        span = spans[0]
        self.assertEqual(span.get_span_context().trace_id, MOCK_W3C_TRACE_ID)

        parent_context = span.parent
        self.assertEqual(
            parent_context.trace_id, span.get_span_context().trace_id
        )
        self.assertEqual(parent_context.span_id, MOCK_W3C_PARENT_SPAN_ID)
        self.assertEqual(len(parent_context.trace_state), 3)
        self.assertEqual(
            parent_context.trace_state.get(MOCK_W3C_TRACE_STATE_KEY),
            MOCK_W3C_TRACE_STATE_VALUE,
        )
        self.assertTrue(parent_context.is_remote)

        test_env_patch.stop()

    def test_using_custom_extractor(self):
        def custom_event_context_extractor(lambda_event):
            return get_global_textmap().extract(lambda_event["foo"]["headers"])

        test_env_patch = mock.patch.dict(
            "os.environ",
            {
                **os.environ,
                # DO NOT use `otel-instrument` script, resort to "manual"
                # instrumentation below
                AWS_LAMBDA_EXEC_WRAPPER: "",
                # NOT Active Tracing
                _X_AMZN_TRACE_ID: MOCK_XRAY_TRACE_CONTEXT_NOT_SAMPLED,
                # NOT using the X-Ray Propagator
                "OTEL_PROPAGATORS": "tracecontext",
            },
        )
        test_env_patch.start()

        # NOTE: Instead of using `AWS_LAMBDA_EXEC_WRAPPER` to point `_HANDLER`
        # to a module which instruments and calls the user `ORIG_HANDLER`, we
        # leave `_HANDLER` as is and replace `AWS_LAMBDA_EXEC_WRAPPER` with this
        # line below. This is like "manual" instrumentation for Lambda.
        AwsLambdaInstrumentor().instrument(
            event_context_extractor=custom_event_context_extractor,
            skip_dep_check=True,
        )

        mock_execute_lambda(
            {
                "foo": {
                    "headers": {
                        "traceparent": MOCK_W3C_TRACE_CONTEXT_SAMPLED,
                        "tracestate": f"{MOCK_W3C_TRACE_STATE_KEY}={MOCK_W3C_TRACE_STATE_VALUE},foo=1,bar=2",
                    }
                }
            }
        )

        spans = self.memory_exporter.get_finished_spans()

        assert spans

        self.assertEqual(len(spans), 1)
        span = spans[0]
        self.assertEqual(span.get_span_context().trace_id, MOCK_W3C_TRACE_ID)

        parent_context = span.parent
        self.assertEqual(
            parent_context.trace_id, span.get_span_context().trace_id
        )
        self.assertEqual(parent_context.span_id, MOCK_W3C_PARENT_SPAN_ID)
        self.assertEqual(len(parent_context.trace_state), 3)
        self.assertEqual(
            parent_context.trace_state.get(MOCK_W3C_TRACE_STATE_KEY),
            MOCK_W3C_TRACE_STATE_VALUE,
        )
        self.assertTrue(parent_context.is_remote)

        test_env_patch.stop()
