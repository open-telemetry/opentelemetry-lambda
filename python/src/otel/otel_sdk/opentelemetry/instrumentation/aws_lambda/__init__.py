# Copyright 2020, OpenTelemetry Authors
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

# TODO: usage
"""
The opentelemetry-instrumentation-aws-lambda package allows tracing AWS
Lambda function.

Usage
-----

.. code:: python
    # Copy this snippet into AWS Lambda function
    # Ref Doc: https://docs.aws.amazon.com/lambda/latest/dg/lambda-python.html

    import boto3
    from opentelemetry.instrumentation.aws_lambda import (
        AwsLambdaInstrumentor
    )

    # Enable instrumentation
    AwsLambdaInstrumentor().instrument(skip_dep_check=True)

    # Lambda function
    def lambda_handler(event, context):
        s3 = boto3.resource('s3')
        for bucket in s3.buckets.all():
            print(bucket.name)

        return "200 OK"

API
---
"""

import logging
import os
from importlib import import_module
from typing import Collection
from wrapt import wrap_function_wrapper

# TODO: aws propagator
from opentelemetry.sdk.extension.aws.trace.propagation.aws_xray_format import (
    AwsXRayFormat,
)
from opentelemetry.instrumentation.aws_lambda.package import _instruments
from opentelemetry.instrumentation.aws_lambda.version import __version__
from opentelemetry.instrumentation.instrumentor import BaseInstrumentor
from opentelemetry.instrumentation.utils import unwrap
from opentelemetry.semconv.trace import SpanAttributes
from opentelemetry.trace import SpanKind, get_tracer, get_tracer_provider

logger = logging.getLogger(__name__)


class AwsLambdaInstrumentor(BaseInstrumentor):
    def instrumentation_dependencies(self) -> Collection[str]:
        return _instruments

    def _instrument(self, **kwargs):
        """Instruments Lambda Handlers on AWS Lambda

        Args:
            **kwargs: Optional arguments
                ``tracer_provider``: a TracerProvider, defaults to global
        """
        tracer = get_tracer(
            __name__, __version__, kwargs.get("tracer_provider")
        )

        lambda_handler = os.environ.get(
            "ORIG_HANDLER", os.environ.get("_HANDLER")
        )
        wrapped_names = lambda_handler.rsplit(".", 1)
        self._wrapped_module_name = wrapped_names[0]
        self._wrapped_function_name = wrapped_names[1]

        _instrument(
            tracer, self._wrapped_module_name, self._wrapped_function_name
        )

    def _uninstrument(self, **kwargs):
        unwrap(
            import_module(self._wrapped_module_name),
            self._wrapped_function_name,
        )


def _instrument(tracer, wrapped_module_name, wrapped_function_name):
    def _instrumented_lambda_handler_call(call_wrapped, instance, args, kwargs):
        orig_handler_name = ".".join(
            [wrapped_module_name, wrapped_function_name]
        )

        # TODO: enable propagate from AWS by env variable
        xray_trace_id = os.environ.get("_X_AMZN_TRACE_ID", "")
        propagator = AwsXRayFormat()
        parent_context = propagator.extract({"X-Amzn-Trace-Id": xray_trace_id})

        with tracer.start_as_current_span(
            name=orig_handler_name, context=parent_context, kind=SpanKind.SERVER
        ) as span:
            if span.is_recording():
                lambda_context = args[1]
                # Refer: https://github.com/open-telemetry/opentelemetry-specification/blob/master/specification/trace/semantic_conventions/faas.md#example
                span.set_attribute(
                    SpanAttributes.FAAS_EXECUTION, lambda_context.aws_request_id
                )
                span.set_attribute(
                    "faas.id", lambda_context.invoked_function_arn
                )

                # TODO: fix in Collector because they belong resource attrubutes
                span.set_attribute(
                    "faas.name", os.environ.get("AWS_LAMBDA_FUNCTION_NAME")
                )
                span.set_attribute(
                    "faas.version",
                    os.environ.get("AWS_LAMBDA_FUNCTION_VERSION"),
                )

            result = call_wrapped(*args, **kwargs)

        # force_flush before function quit in case of Lambda freeze.
        tracer_provider = get_tracer_provider()
        tracer_provider.force_flush()

        return result

    wrap_function_wrapper(
        wrapped_module_name,
        wrapped_function_name,
        _instrumented_lambda_handler_call,
    )
