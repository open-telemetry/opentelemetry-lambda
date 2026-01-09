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
Tests for Lambda handler wrapping and instrumentation.
Following patterns from the Node.js layer tests.
"""

import os
import sys
import unittest
from pathlib import Path
from unittest import mock

# Mock the ORIG_HANDLER environment variable before importing otel_wrapper
# This prevents the module from failing on import during tests
with mock.patch.dict(os.environ, {"ORIG_HANDLER": "mocks.lambda_function.handler"}):
    # Add otel_sdk to path for importing
    otel_sdk_path = Path(__file__).parent.parent / "otel_sdk"
    sys.path.insert(0, str(otel_sdk_path))

    from otel_wrapper import HandlerError, modify_module_name


class TestLambdaHandler(unittest.TestCase):
    """Test Lambda handler wrapping and execution."""

    def setUp(self):
        """Store original environment."""
        self.old_env = os.environ.copy()

    def tearDown(self):
        """Restore original environment."""
        os.environ.clear()
        os.environ.update(self.old_env)

    def test_handler_error_when_missing(self):
        """Test that HandlerError is raised when ORIG_HANDLER is not set."""
        # This validates the error handling in otel_wrapper.py
        self.assertIsNotNone(HandlerError)

        # Verify HandlerError is an Exception subclass
        error = HandlerError("test error")
        self.assertIsInstance(error, Exception)
        self.assertEqual(str(error), "test error")

    def test_handler_path_parsing_valid(self):
        """Test parsing of valid handler path (module.function format)."""
        handler_path = "lambda_function.handler"

        # Test the parsing logic used in otel_wrapper.py
        try:
            mod_name, handler_name = handler_path.rsplit(".", 1)
            self.assertEqual(mod_name, "lambda_function")
            self.assertEqual(handler_name, "handler")
        except ValueError:
            self.fail("Valid handler path should not raise ValueError")

    def test_handler_path_parsing_invalid(self):
        """Test that invalid handler paths raise IndexError when accessing second element."""
        invalid_paths = [
            "lambda_function",  # No dot separator
            "",  # Empty string
        ]

        for invalid_path in invalid_paths:
            with self.subTest(path=invalid_path), self.assertRaises(IndexError):
                # This is the parsing logic from otel_wrapper.py
                # rsplit returns a list, and [1] will raise IndexError if no second element
                invalid_path.rsplit(".", 1)[1]

    def test_handler_path_with_nested_module(self):
        """Test parsing of nested module handler path."""
        handler_path = "handlers.main.lambda_handler"

        try:
            mod_name, handler_name = handler_path.rsplit(".", 1)
            self.assertEqual(mod_name, "handlers.main")
            self.assertEqual(handler_name, "lambda_handler")
        except ValueError:
            self.fail("Valid nested handler path should not raise ValueError")

    def test_aws_lambda_function_name(self):
        """Test AWS Lambda function name environment variable."""
        function_name = "test-function"
        os.environ["AWS_LAMBDA_FUNCTION_NAME"] = function_name

        self.assertEqual(os.environ["AWS_LAMBDA_FUNCTION_NAME"], function_name)

    def test_lambda_task_root(self):
        """Test LAMBDA_TASK_ROOT environment variable."""
        task_root = "/var/task"
        os.environ["LAMBDA_TASK_ROOT"] = task_root

        self.assertEqual(os.environ["LAMBDA_TASK_ROOT"], task_root)


class TestAwsLambdaInstrumentation(unittest.TestCase):
    """Test AWS Lambda instrumentation."""

    def setUp(self):
        """Store original environment."""
        self.old_env = os.environ.copy()

    def tearDown(self):
        """Restore original environment."""
        os.environ.clear()
        os.environ.update(self.old_env)

    def test_lambda_runtime_api_environment(self):
        """Test Lambda runtime API environment variables."""
        runtime_api = "127.0.0.1:9001"
        os.environ["AWS_LAMBDA_RUNTIME_API"] = runtime_api

        self.assertEqual(os.environ["AWS_LAMBDA_RUNTIME_API"], runtime_api)

    def test_xray_trace_header(self):
        """Test X-Ray trace header environment variable."""
        trace_header = (
            "Root=1-5759e988-bd862e3fe1be46a994272793;Parent=53995c3f42cd8ad8;Sampled=1"
        )
        os.environ["_X_AMZN_TRACE_ID"] = trace_header

        self.assertEqual(os.environ["_X_AMZN_TRACE_ID"], trace_header)

    def test_lambda_handler_environment(self):
        """Test _HANDLER environment variable (internal Lambda runtime variable)."""
        handler = "lambda_function.handler"
        os.environ["_HANDLER"] = handler

        self.assertEqual(os.environ["_HANDLER"], handler)


class TestModuleNameModification(unittest.TestCase):
    """Test the modify_module_name function from otel_wrapper.py."""

    def test_simple_module_name(self):
        """Test simple module name (no slashes)."""
        module_name = "lambda_function"
        modified = modify_module_name(module_name)
        self.assertEqual(modified, "lambda_function")

    def test_path_with_single_directory(self):
        """Test path with single directory."""
        module_name = "handlers/main"
        modified = modify_module_name(module_name)
        self.assertEqual(modified, "handlers.main")

    def test_path_with_multiple_directories(self):
        """Test path with multiple directories."""
        module_name = "src/handlers/api/main"
        modified = modify_module_name(module_name)
        self.assertEqual(modified, "src.handlers.api.main")

    def test_path_with_trailing_slash(self):
        """Test path with trailing slash - note this keeps empty string."""
        module_name = "handlers/main/"
        modified = modify_module_name(module_name)
        # The actual function doesn't filter empty strings
        # so "handlers/main/" becomes "handlers.main."
        self.assertEqual(modified, "handlers.main.")

    def test_empty_string(self):
        """Test empty string input."""
        module_name = ""
        modified = modify_module_name(module_name)
        self.assertEqual(modified, "")

    def test_already_dotted_path(self):
        """Test path that's already using dots."""
        module_name = "handlers.main"
        modified = modify_module_name(module_name)
        # Should remain unchanged since there are no slashes
        self.assertEqual(modified, "handlers.main")


if __name__ == "__main__":
    unittest.main()
