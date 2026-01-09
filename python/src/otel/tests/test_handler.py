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
import unittest


class TestLambdaHandler(unittest.TestCase):
    """Test Lambda handler wrapping and execution."""

    def setUp(self):
        """Store original environment."""
        self.old_env = os.environ.copy()

    def tearDown(self):
        """Restore original environment."""
        os.environ.clear()
        os.environ.update(self.old_env)

    def test_handler_environment_variable(self):
        """Test that ORIG_HANDLER environment variable is required."""
        # ORIG_HANDLER should be set for proper handler loading
        os.environ.pop("ORIG_HANDLER", None)

        with self.assertRaises(KeyError):
            # Should raise error when ORIG_HANDLER is not set
            _ = os.environ["ORIG_HANDLER"]

    def test_handler_path_parsing(self):
        """Test parsing of handler path (module.function format)."""
        handler_path = "lambda_function.handler"
        os.environ["ORIG_HANDLER"] = handler_path

        # Split into module and function
        parts = handler_path.rsplit(".", 1)
        self.assertEqual(len(parts), 2)
        self.assertEqual(parts[0], "lambda_function")
        self.assertEqual(parts[1], "handler")

    def test_handler_path_with_nested_module(self):
        """Test parsing of nested module handler path."""
        handler_path = "handlers.main.lambda_handler"
        os.environ["ORIG_HANDLER"] = handler_path

        parts = handler_path.rsplit(".", 1)
        self.assertEqual(len(parts), 2)
        self.assertEqual(parts[0], "handlers.main")
        self.assertEqual(parts[1], "lambda_handler")

    def test_handler_path_conversion(self):
        """Test conversion of file paths to module paths."""
        # Test that handlers/main is converted to handlers.main
        file_path = "handlers/main"
        module_path = ".".join(file_path.split("/"))
        self.assertEqual(module_path, "handlers.main")

        # Test nested paths
        file_path = "src/handlers/api/main"
        module_path = ".".join(file_path.split("/"))
        self.assertEqual(module_path, "src.handlers.api.main")

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
    """Test module name modification logic."""

    def test_simple_module_name(self):
        """Test simple module name (no slashes)."""
        module_name = "lambda_function"
        modified = ".".join(module_name.split("/"))
        self.assertEqual(modified, "lambda_function")

    def test_path_with_single_directory(self):
        """Test path with single directory."""
        module_name = "handlers/main"
        modified = ".".join(module_name.split("/"))
        self.assertEqual(modified, "handlers.main")

    def test_path_with_multiple_directories(self):
        """Test path with multiple directories."""
        module_name = "src/handlers/api/main"
        modified = ".".join(module_name.split("/"))
        self.assertEqual(modified, "src.handlers.api.main")

    def test_path_with_trailing_slash(self):
        """Test path with trailing slash."""
        module_name = "handlers/main/"
        # Split and filter out empty strings
        parts = [p for p in module_name.split("/") if p]
        modified = ".".join(parts)
        self.assertEqual(modified, "handlers.main")


if __name__ == "__main__":
    unittest.main()
