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

"""
This file tests that the `otel-instrument` script included in this repository
successfully instruments OTel Python in a mock Lambda environment.
"""

import fileinput
import os
import subprocess
import sys
from importlib import import_module, reload
from shutil import which
from unittest import mock
from opentelemetry import propagate
from opentelemetry.environment_variables import OTEL_PROPAGATORS
from opentelemetry.instrumentation.aws_lambda import (
    _HANDLER,
    _X_AMZN_TRACE_ID,
    ORIG_HANDLER,
    AwsLambdaInstrumentor,
)
from opentelemetry.propagators.aws.aws_xray_propagator import (
    TRACE_ID_FIRST_PART_LENGTH,
    TRACE_ID_VERSION,
)
from opentelemetry.semconv.resource import ResourceAttributes
from opentelemetry.semconv.trace import SpanAttributes
from opentelemetry.test.test_base import TestBase
from opentelemetry.trace import SpanKind
from opentelemetry.trace.propagation.tracecontext import (
    TraceContextTextMapPropagator,
)

AWS_LAMBDA_EXEC_WRAPPER = "AWS_LAMBDA_EXEC_WRAPPER"
INIT_OTEL_SCRIPTS_DIR = os.path.join(
    *(os.path.dirname(__file__), "..", "otel_sdk")
)
TOX_PYTHON_DIRECTORY = os.path.dirname(os.path.dirname(which("python3")))


class MockLambdaContext:
    def __init__(self, aws_request_id, invoked_function_arn):
        self.invoked_function_arn = invoked_function_arn
        self.aws_request_id = aws_request_id


MOCK_LAMBDA_CONTEXT = MockLambdaContext(
    aws_request_id="mock_aws_request_id",
    invoked_function_arn="arn:aws:lambda:us-west-2:123456789012:function:my-function",
)

MOCK_XRAY_TRACE_ID = 0x5FB7331105E8BB83207FA31D4D9CDB4C
MOCK_XRAY_TRACE_ID_STR = f"{MOCK_XRAY_TRACE_ID:x}"
MOCK_XRAY_PARENT_SPAN_ID = 0x3328B8445A6DBAD2
MOCK_XRAY_TRACE_CONTEXT_COMMON = f"Root={TRACE_ID_VERSION}-{MOCK_XRAY_TRACE_ID_STR[:TRACE_ID_FIRST_PART_LENGTH]}-{MOCK_XRAY_TRACE_ID_STR[TRACE_ID_FIRST_PART_LENGTH:]};Parent={MOCK_XRAY_PARENT_SPAN_ID:x}"
MOCK_XRAY_TRACE_CONTEXT_SAMPLED = f"{MOCK_XRAY_TRACE_CONTEXT_COMMON};Sampled=1"
MOCK_XRAY_TRACE_CONTEXT_NOT_SAMPLED = (
    f"{MOCK_XRAY_TRACE_CONTEXT_COMMON};Sampled=0"
)

# See more:
# https://www.w3.org/TR/trace-context/#examples-of-http-traceparent-headers

MOCK_W3C_TRACE_ID = 0x5CE0E9A56015FEC5AADFA328AE398115
MOCK_W3C_PARENT_SPAN_ID = 0xAB54A98CEB1F0AD2
MOCK_W3C_TRACE_CONTEXT_SAMPLED = (
    f"00-{MOCK_W3C_TRACE_ID:x}-{MOCK_W3C_PARENT_SPAN_ID:x}-01"
)

MOCK_W3C_TRACE_STATE_KEY = "vendor_specific_key"
MOCK_W3C_TRACE_STATE_VALUE = "test_value"


def replace_in_file(filename, old_text, new_text):
    with fileinput.FileInput(filename, inplace=True) as file_object:
        for line in file_object:
            # This directs the output to the file, not the console
            print(line.replace(old_text, new_text), end="")


def mock_aws_lambda_exec_wrapper():
    """Mocks automatically instrumenting user Lambda function by pointing
    `AWS_LAMBDA_EXEC_WRAPPER` to the `otel-instrument` script.

    TODO: It would be better if `moto`'s `mock_lambda` supported setting
    AWS_LAMBDA_EXEC_WRAPPER so we could make the call to Lambda instead.

    See more:
    https://aws-otel.github.io/docs/getting-started/lambda/lambda-python
    """

    # NOTE: Because we run as a subprocess, the python packages are NOT patched
    # with instrumentation. In this test we just make sure we can complete auto
    # instrumentation without error and the correct environment variabels are
    # set. A future improvement might have us run `opentelemetry-instrument` in
    # this process to imitate `otel-instrument`, but our lambda handler does not
    # call other instrumented libraries so we have no use for it for now.

    print_environ_program = (
        "import os;"
        f"print(f\"{ORIG_HANDLER}={{os.environ['{ORIG_HANDLER}']}}\");"
        f"print(f\"{_HANDLER}={{os.environ['{_HANDLER}']}}\");"
    )

    completed_subprocess = subprocess.run(
        [
            os.path.join(INIT_OTEL_SCRIPTS_DIR, "otel-instrument"),
            "python3",
            "-c",
            print_environ_program,
        ],
        check=True,
        stdout=subprocess.PIPE,
        text=True,
    )

    # NOTE: Because `otel-instrument` cannot affect this python environment, we
    # parse the stdout produced by our test python program to update the
    # environment in this parent python process.

    for env_var_line in completed_subprocess.stdout.split("\n"):
        if env_var_line:
            env_key, env_value = env_var_line.split("=")
            os.environ[env_key] = env_value


def mock_execute_lambda(event=None):
    """Mocks the AWS Lambda execution. Mocks importing and then calling the
    method at the current `_HANDLER` environment variable. Like the real Lambda,
    if `AWS_LAMBDA_EXEC_WRAPPER` is defined, it executes that before `_HANDLER`.

    NOTE: We don't use `moto`'s `mock_lambda` because we are not instrumenting
    calls to AWS Lambda using the AWS SDK. Instead, we are instrumenting AWS
    Lambda itself.

    See more:
    https://docs.aws.amazon.com/lambda/latest/dg/runtimes-modify.html#runtime-wrapper

    Args:
        event: The Lambda event which may or may not be used by instrumentation.
    """

    # The point of the repo is to test using the script, so we can count on it
    # being here for every test and do not check for its existence.
    # if os.environ[AWS_LAMBDA_EXEC_WRAPPER]:
    globals()[os.environ[AWS_LAMBDA_EXEC_WRAPPER]]()

    module_name, handler_name = os.environ[_HANDLER].rsplit(".", 1)
    handler_module = import_module(module_name.replace("/", "."))
    getattr(handler_module, handler_name)(event, MOCK_LAMBDA_CONTEXT)


class TestAwsLambdaInstrumentor(TestBase):
    """AWS Lambda Instrumentation Testsuite"""

    @classmethod
    def setUpClass(cls):
        super().setUpClass()
        sys.path.append(INIT_OTEL_SCRIPTS_DIR)
        replace_in_file(
            os.path.join(INIT_OTEL_SCRIPTS_DIR, "otel-instrument"),
            'export LAMBDA_LAYER_PKGS_DIR="/opt/python"',
            f'export LAMBDA_LAYER_PKGS_DIR="{TOX_PYTHON_DIRECTORY}"',
        )

    def setUp(self):
        super().setUp()
        self.common_env_patch = mock.patch.dict(
            "os.environ",
            {
                AWS_LAMBDA_EXEC_WRAPPER: "mock_aws_lambda_exec_wrapper",
                _HANDLER: "mocks.lambda_function.handler",
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
        sys.path.remove(INIT_OTEL_SCRIPTS_DIR)
        replace_in_file(
            os.path.join(INIT_OTEL_SCRIPTS_DIR, "otel-instrument"),
            f'export LAMBDA_LAYER_PKGS_DIR="{TOX_PYTHON_DIRECTORY}"',
            'export LAMBDA_LAYER_PKGS_DIR="/opt/python"',
        )

    def test_active_tracing(self):
        test_env_patch = mock.patch.dict(
            "os.environ",
            {
                **os.environ,
                # Using Active tracing
                _X_AMZN_TRACE_ID: MOCK_XRAY_TRACE_CONTEXT_SAMPLED,
                OTEL_PROPAGATORS: "xray-lambda"
            },
        )
        test_env_patch.start()

        # try to load propagators based on the OTEL_PROPAGATORS env var
        reload(propagate)

        mock_execute_lambda()

        spans = self.memory_exporter.get_finished_spans()

        assert spans

        self.assertEqual(len(spans), 1)
        span = spans[0]
        self.assertEqual(span.name, os.environ[ORIG_HANDLER])
        self.assertEqual(span.get_span_context().trace_id, MOCK_XRAY_TRACE_ID)
        self.assertEqual(span.kind, SpanKind.SERVER)
        self.assertSpanHasAttributes(
            span,
            {
                ResourceAttributes.CLOUD_RESOURCE_ID: MOCK_LAMBDA_CONTEXT.invoked_function_arn,
                SpanAttributes.FAAS_INVOCATION_ID: MOCK_LAMBDA_CONTEXT.aws_request_id,
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
                OTEL_PROPAGATORS: "tracecontext",
            },
        )
        test_env_patch.start()

        # try to load propagators based on the OTEL_PROPAGATORS env var
        reload(propagate)

        mock_execute_lambda(
            {
                "headers": {
                    TraceContextTextMapPropagator._TRACEPARENT_HEADER_NAME: MOCK_W3C_TRACE_CONTEXT_SAMPLED,
                    TraceContextTextMapPropagator._TRACESTATE_HEADER_NAME: f"{MOCK_W3C_TRACE_STATE_KEY}={MOCK_W3C_TRACE_STATE_VALUE},foo=1,bar=2",
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
