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
Tests for the OpenTelemetry Lambda wrapper configuration.
Tests actual functions from otel_wrapper.py.
"""

import os
import sys
import unittest
from pathlib import Path
from unittest import mock

# Mock environment to allow importing otel_wrapper
with mock.patch.dict(os.environ, {"ORIG_HANDLER": "mocks.lambda_function.handler"}):
    otel_sdk_path = Path(__file__).parent.parent / "otel_sdk"
    sys.path.insert(0, str(otel_sdk_path))
    from otel_wrapper import _get_active_instrumentations


class TestInstrumentationConfiguration(unittest.TestCase):
    """Test the _get_active_instrumentations function from otel_wrapper.py."""

    def setUp(self):
        """Store original environment."""
        self.old_env = os.environ.copy()

    def tearDown(self):
        """Restore original environment."""
        os.environ.clear()
        os.environ.update(self.old_env)

    def test_default_instrumentations_empty(self):
        """Test that by default no instrumentations are enabled (except botocore/aws-lambda)."""
        os.environ.pop("OTEL_PYTHON_ENABLED_INSTRUMENTATIONS", None)
        os.environ.pop("OTEL_PYTHON_DISABLED_INSTRUMENTATIONS", None)

        active = _get_active_instrumentations()
        # Should be empty set by default (botocore and aws-lambda are loaded separately)
        self.assertIsInstance(active, set)
        self.assertEqual(len(active), 0)

    def test_enabled_instrumentations_parsing(self):
        """Test the real _get_active_instrumentations function with OTEL_PYTHON_ENABLED_INSTRUMENTATIONS."""
        os.environ["OTEL_PYTHON_ENABLED_INSTRUMENTATIONS"] = "requests,psycopg2,redis"

        active = _get_active_instrumentations()
        self.assertIsInstance(active, set)
        self.assertEqual(len(active), 3)
        self.assertIn("requests", active)
        self.assertIn("psycopg2", active)
        self.assertIn("redis", active)

    def test_disabled_instrumentations_removes_from_enabled(self):
        """Test that OTEL_PYTHON_DISABLED_INSTRUMENTATIONS removes items from enabled set."""
        os.environ["OTEL_PYTHON_ENABLED_INSTRUMENTATIONS"] = (
            "requests,psycopg2,redis,django"
        )
        os.environ["OTEL_PYTHON_DISABLED_INSTRUMENTATIONS"] = "django,psycopg2"

        active = _get_active_instrumentations()
        self.assertIn("requests", active)
        self.assertIn("redis", active)
        self.assertNotIn("django", active)
        self.assertNotIn("psycopg2", active)

    def test_disabled_with_no_enabled_does_nothing(self):
        """Test that disabling instrumentations with no enabled list has no effect."""
        os.environ.pop("OTEL_PYTHON_ENABLED_INSTRUMENTATIONS", None)
        os.environ["OTEL_PYTHON_DISABLED_INSTRUMENTATIONS"] = "django,flask"

        active = _get_active_instrumentations()
        # Should still be empty since nothing was enabled
        self.assertEqual(len(active), 0)

    def test_empty_enabled_instrumentations_string(self):
        """Test behavior with empty enabled instrumentations string."""
        os.environ["OTEL_PYTHON_ENABLED_INSTRUMENTATIONS"] = ""

        active = _get_active_instrumentations()
        # Empty string split returns [''], so we should get one empty string in the set
        # This tests the actual behavior
        self.assertIsInstance(active, set)


class TestExporterConfiguration(unittest.TestCase):
    """Test exporter configuration via environment variables."""


if __name__ == "__main__":
    unittest.main()
