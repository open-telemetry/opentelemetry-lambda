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
from typing import Any, Collection

from opentelemetry.context.context import Context
from opentelemetry.instrumentation.aws_lambda.package import _instruments
from opentelemetry.instrumentation.aws_lambda.version import __version__
from opentelemetry.instrumentation.instrumentor import BaseInstrumentor
from opentelemetry.instrumentation.utils import unwrap
from opentelemetry.propagate import get_global_textmap
from opentelemetry.sdk.extension.aws.trace.propagation.aws_xray_format import (
    TRACE_HEADER_KEY,
    AwsXRayFormat,
)
from opentelemetry.semconv.trace import SpanAttributes
from opentelemetry.trace import (
    SpanKind,
    Tracer,
    get_tracer,
    get_tracer_provider,
)
from opentelemetry.trace.propagation import get_current_span
from wrapt import wrap_function_wrapper

logger = logging.getLogger(__name__)


class AwsLambdaInstrumentor(BaseInstrumentor):
    def instrumentation_dependencies(self) -> Collection[str]:
        return _instruments

    def _instrument(self, **kwargs):
        """Instruments Lambda Handlers on AWS Lambda.

        See more:
        https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/trace/semantic_conventions/instrumentation/aws-lambda.md#instrumenting-aws-lambda

        Args:
            **kwargs: Optional arguments
                ``tracer_provider``: a TracerProvider, defaults to global
                ``event_context_extractor``: a method which takes the Lambda
                    Event as input and extracts an OTel Context from it. By default,
                    the context is extracted from the HTTP headers of an API Gateway
                    request.
        """
        tracer = get_tracer(
            __name__, __version__, kwargs.get("tracer_provider")
        )
        event_context_extractor = kwargs.get("event_context_extractor")

        lambda_handler = os.environ.get(
            "ORIG_HANDLER", os.environ.get("_HANDLER")
        )
        wrapped_names = lambda_handler.rsplit(".", 1)
        self._wrapped_module_name = wrapped_names[0]
        self._wrapped_function_name = wrapped_names[1]

        _instrument(
            tracer,
            self._wrapped_module_name,
            self._wrapped_function_name,
            event_context_extractor,
        )

    def _uninstrument(self, **kwargs):
        unwrap(
            import_module(self._wrapped_module_name),
            self._wrapped_function_name,
        )


def _default_event_context_extractor(lambda_event: Any) -> Context:
    """Default way of extracting the context from the Lambda Event.

    Assumes the Lambda Event is a map with the headers under the 'headers' key.
    This is the mapping to use when the Lambda is invoked by an API Gateway
    REST API where API Gateway is acting as a pure proxy for the request.

    See more:
    https://docs.aws.amazon.com/apigateway/latest/developerguide/set-up-lambda-proxy-integrations.html#api-gateway-simple-proxy-for-lambda-input-format

    Args:
        lambda_event: user-defined, so it could be anything, but this
            method counts it being a map with a 'headers' key
    Returns:
        A Context with configuration found in the carrier.
    """
    try:
        headers = lambda_event["headers"]
    except (TypeError, KeyError):
        logger.debug(
            "Extracting context from Lambda Event failed: either enable X-Ray active tracing or configure API Gateway to trigger this Lambda function as a pure proxy. Otherwise, generated spans will have an invalid (empty) parent context."
        )
        headers = {}
    return get_global_textmap().extract(headers)


def _instrument(
    tracer: Tracer,
    wrapped_module_name,
    wrapped_function_name,
    event_context_extractor=None,
):
    def _determine_parent_context(lambda_event: Any) -> Context:
        """Determine the parent context for the current Lambda invocation.

        See more:
        https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/trace/semantic_conventions/instrumentation/aws-lambda.md#determining-the-parent-of-a-span

        Args:
            lambda_event: user-defined, so it could be anything, but this
                method counts it being a map with a 'headers' key
        Returns:
            A Context with configuration found in the carrier.
        """
        parent_context = None

        xray_env_var = os.environ.get("_X_AMZN_TRACE_ID")

        if xray_env_var:
            parent_context = AwsXRayFormat().extract(
                {TRACE_HEADER_KEY: xray_env_var}
            )

        if (
            parent_context
            and get_current_span(parent_context)
            .get_span_context()
            .trace_flags.sampled
        ):
            return parent_context

        if event_context_extractor:
            parent_context = event_context_extractor(lambda_event)
        else:
            parent_context = _default_event_context_extractor(lambda_event)

        return parent_context

    def _instrumented_lambda_handler_call(call_wrapped, instance, args, kwargs):
        orig_handler_name = ".".join(
            [wrapped_module_name, wrapped_function_name]
        )

        lambda_event = args[0]

        parent_context = _determine_parent_context(lambda_event)

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
