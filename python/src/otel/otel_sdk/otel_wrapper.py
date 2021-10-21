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
    from opentelemetry.instrumentation.aws_lambda import otel_handler

    # Lambda function
    @otel_handler
    def lambda_handler(event, context):
        s3 = boto3.resource('s3')
        for bucket in s3.buckets.all():
            print(bucket.name)

        return "200 OK"

API
---
"""


import os
from importlib import import_module

from opentelemetry.instrumentation.aws_lambda import otel_handler


wrapped_handler = None


def lambda_handler(*args, **kwargs):
    global wrapped_handler
    if wrapped_handler is None:
        lambda_handler = os.environ.get("OTEL_INSTRUMENTATION_AWS_LAMBDA_HANDLER")
        module_name, func_name = lambda_handler.rsplit('.', 1)
        module = import_module(module_name)
        func = getattr(module, func_name)
        wrapped_handler = otel_handler(func)
    return wrapped_handler(*args, **kwargs)
