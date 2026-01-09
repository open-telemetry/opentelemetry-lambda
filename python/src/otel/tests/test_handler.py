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
    """Test Lambda handler functions from otel_wrapper.py."""

    def setUp(self):
        """Store original environment."""
        self.old_env = os.environ.copy()

    def tearDown(self):
        """Restore original environment."""
        os.environ.clear()
        os.environ.update(self.old_env)

    def test_handler_error_class_exists(self):
        """Test that HandlerError class is properly defined."""
        # This validates the error handling in otel_wrapper.py
        self.assertIsNotNone(HandlerError)

        # Verify HandlerError is an Exception subclass
        error = HandlerError("test error")
        self.assertIsInstance(error, Exception)
        self.assertEqual(str(error), "test error")

    def test_handler_error_message_formatting(self):
        """Test HandlerError message formatting."""
        error = HandlerError("ORIG_HANDLER is not defined.")
        self.assertEqual(str(error), "ORIG_HANDLER is not defined.")

        error2 = HandlerError("Bad path 'invalid' for ORIG_HANDLER: no dot found")
        self.assertIn("Bad path", str(error2))
        self.assertIn("ORIG_HANDLER", str(error2))


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
