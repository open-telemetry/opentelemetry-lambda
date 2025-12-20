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
OpenTelemetry Lambda Handler Wrapper

This module wraps the user's Lambda function to enable automatic OpenTelemetry
instrumentation. It acts as wrapper script that instruments the Lambda function
before loading the module containing the user's handler.

The instrumentation process works as follows:
------------
1. The `otel-instrument` shell script sets _HANDLER to point to this file's
   `lambda_handler`, saving the original handler path to ORIG_HANDLER.

2. When AWS Lambda imports this module, `auto_instrumentation.initialize()` runs
   immediately, instrumenting the application before any user code executes.

3. The module containing the user's handler is loaded by this script and the
  `lambda_handler` variable is bound to the user's original handler function,
   allowing Lambda invocations to be transparently forwarded to the original handler.

Details on why the `opentelemetry-instrument` CLI wrapper is insufficient:
------------------------------------------------
The `opentelemetry-instrument` CLI wrapper only instruments the initial Python process.
AWS Lambda may spawn fresh Python processes for new invocations (e.g. as is the case with
lambda managed instances), which would bypass CLI based instrumentation. By
calling `auto_instrumentation.initialize()` at module import time, we ensure every
Lambda execution context is instrumented.

Environment Variables
---------------------
ORIG_HANDLER : str
    The original Lambda handler path (e.g., "mymodule.handler"). Set by
    `otel-instrument` before this module is loaded.
"""


import os
from importlib import import_module

from opentelemetry.instrumentation import auto_instrumentation

# Initialize OpenTelemetry instrumentation immediately on module import.
# This must happen before the user's handler module is loaded (below) to ensure
# all library patches are applied before any user code runs.
auto_instrumentation.initialize()


def _get_orig_handler():
    """
    Resolve and return the user's original Lambda handler function.

    Reads the handler path from the ORIG_HANDLER environment variable,
    dynamically imports the handler's module and returns the handler
    function.
    """

    handler_path = os.environ.get("ORIG_HANDLER")

    if handler_path is None:
        raise RuntimeError(
            "ORIG_HANDLER is not defined."
        )

    # Split "module/path.handler_name" into module path and function name.
    # The handler path uses the last "." as the separator between module and function.
    try:
        module_path, handler_name = handler_path.rsplit(".", 1)
    except ValueError as e:
        raise RuntimeError(
            f"Invalid ORIG_HANDLER format '{handler_path}': expected "
            f"'module.handler_name' or 'path/to/module.handler_name'. Error: {e}"
        )

    # Convert path separators to Python module notation
    module_name = ".".join(module_path.split("/"))

    handler_module = import_module(module_name)
    return getattr(handler_module, handler_name)


# Resolve to the user's handler at module load time.
lambda_handler = _get_orig_handler()
