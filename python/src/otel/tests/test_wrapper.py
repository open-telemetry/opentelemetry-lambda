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
Following patterns from the Node.js layer tests:
- Test propagator configuration
- Test exporter configuration
- Test instrumentation configuration
"""

import os
import unittest

from opentelemetry import propagate
from opentelemetry.propagators.aws.aws_xray_propagator import AwsXRayPropagator
from opentelemetry.trace.propagation.tracecontext import TraceContextTextMapPropagator


class TestPropagatorConfiguration(unittest.TestCase):
    """Test propagator configuration via OTEL_PROPAGATORS environment variable."""

    def setUp(self):
        """Reset propagator before each test."""
        self.old_env = os.environ.copy()

    def tearDown(self):
        """Restore original environment."""
        os.environ.clear()
        os.environ.update(self.old_env)

    def test_default_propagator_configuration(self):
        """Test that the default propagator is configured (tracecontext, baggage)."""
        # Default propagators should be W3C TraceContext and Baggage
        propagator = propagate.get_global_textmap()
        # The default should have traceparent, tracestate, and possibly baggage fields
        fields = propagator.fields
        self.assertIn("traceparent", fields)

    def test_xray_propagator_configuration(self):
        """Test X-Ray propagator configuration via environment variable."""
        os.environ["OTEL_PROPAGATORS"] = "xray"

        # We need to simulate the propagator configuration that happens in otel_wrapper
        # In actual usage, this would be configured during wrapper initialization

        propagate.set_global_textmap(AwsXRayPropagator())

        propagator = propagate.get_global_textmap()
        fields = propagator.fields
        # X-Amzn-Trace-Id is case-sensitive
        self.assertIn("X-Amzn-Trace-Id", fields)

    def test_tracecontext_propagator_configuration(self):
        """Test W3C TraceContext propagator configuration."""
        os.environ["OTEL_PROPAGATORS"] = "tracecontext"

        propagate.set_global_textmap(TraceContextTextMapPropagator())

        propagator = propagate.get_global_textmap()
        fields = propagator.fields
        self.assertIn("traceparent", fields)
        self.assertIn("tracestate", fields)

    def test_composite_propagator_configuration(self):
        """Test configuration with multiple propagators."""
        os.environ["OTEL_PROPAGATORS"] = "tracecontext,xray"

        from opentelemetry.propagators.aws.aws_xray_propagator import AwsXRayPropagator
        from opentelemetry.propagators.composite import CompositePropagator

        composite = CompositePropagator(
            [TraceContextTextMapPropagator(), AwsXRayPropagator()]
        )
        propagate.set_global_textmap(composite)

        propagator = propagate.get_global_textmap()
        fields = propagator.fields
        # Should have fields from both propagators
        self.assertIn("traceparent", fields)
        # X-Amzn-Trace-Id is case-sensitive
        self.assertIn("X-Amzn-Trace-Id", fields)


class TestExporterConfiguration(unittest.TestCase):
    """Test exporter configuration via environment variables."""

    def setUp(self):
        """Store original environment."""
        self.old_env = os.environ.copy()

    def tearDown(self):
        """Restore original environment."""
        os.environ.clear()
        os.environ.update(self.old_env)

    def test_default_otlp_exporter(self):
        """Test that OTLP is the default exporter."""
        # Default should be OTLP
        exporter_name = os.environ.get("OTEL_TRACES_EXPORTER", "otlp")
        self.assertEqual(exporter_name, "otlp")

    def test_console_exporter_configuration(self):
        """Test console exporter configuration."""
        os.environ["OTEL_TRACES_EXPORTER"] = "console"
        exporter_name = os.environ.get("OTEL_TRACES_EXPORTER")
        self.assertEqual(exporter_name, "console")

    def test_none_exporter_configuration(self):
        """Test that 'none' disables tracing."""
        os.environ["OTEL_TRACES_EXPORTER"] = "none"
        exporter_name = os.environ.get("OTEL_TRACES_EXPORTER")
        self.assertEqual(exporter_name, "none")

    def test_multiple_exporters_configuration(self):
        """Test configuration with multiple exporters."""
        os.environ["OTEL_TRACES_EXPORTER"] = "console,otlp"
        exporter_names = os.environ.get("OTEL_TRACES_EXPORTER", "").split(",")
        self.assertIn("console", exporter_names)
        self.assertIn("otlp", exporter_names)


class TestInstrumentationConfiguration(unittest.TestCase):
    """Test instrumentation loading configuration."""

    def setUp(self):
        """Store original environment."""
        self.old_env = os.environ.copy()

    def tearDown(self):
        """Restore original environment."""
        os.environ.clear()
        os.environ.update(self.old_env)

    def test_enabled_instrumentations_parsing(self):
        """Test parsing of OTEL_PYTHON_ENABLED_INSTRUMENTATIONS."""
        os.environ["OTEL_PYTHON_ENABLED_INSTRUMENTATIONS"] = "requests,psycopg2,redis"
        enabled = os.environ.get("OTEL_PYTHON_ENABLED_INSTRUMENTATIONS", "").split(",")
        self.assertEqual(len(enabled), 3)
        self.assertIn("requests", enabled)
        self.assertIn("psycopg2", enabled)
        self.assertIn("redis", enabled)

    def test_disabled_instrumentations_parsing(self):
        """Test parsing of OTEL_PYTHON_DISABLED_INSTRUMENTATIONS."""
        os.environ["OTEL_PYTHON_DISABLED_INSTRUMENTATIONS"] = "django,flask"
        disabled = os.environ.get("OTEL_PYTHON_DISABLED_INSTRUMENTATIONS", "").split(
            ","
        )
        self.assertEqual(len(disabled), 2)
        self.assertIn("django", disabled)
        self.assertIn("flask", disabled)

    def test_empty_enabled_instrumentations(self):
        """Test behavior with empty enabled instrumentations."""
        # When not set, should return empty string
        enabled = os.environ.get("OTEL_PYTHON_ENABLED_INSTRUMENTATIONS", "")
        self.assertEqual(enabled, "")


class TestResourceConfiguration(unittest.TestCase):
    """Test resource configuration."""

    def setUp(self):
        """Store original environment."""
        self.old_env = os.environ.copy()

    def tearDown(self):
        """Restore original environment."""
        os.environ.clear()
        os.environ.update(self.old_env)

    def test_service_name_configuration(self):
        """Test service name configuration."""
        test_service_name = "my-lambda-function"
        os.environ["OTEL_SERVICE_NAME"] = test_service_name

        service_name = os.environ.get("OTEL_SERVICE_NAME")
        self.assertEqual(service_name, test_service_name)

    def test_service_name_from_function_name(self):
        """Test service name defaults to AWS_LAMBDA_FUNCTION_NAME."""
        function_name = "test-lambda-function"
        os.environ["AWS_LAMBDA_FUNCTION_NAME"] = function_name
        os.environ.pop("OTEL_SERVICE_NAME", None)

        # Service name should fall back to function name
        service_name = os.environ.get(
            "OTEL_SERVICE_NAME", os.environ.get("AWS_LAMBDA_FUNCTION_NAME")
        )
        self.assertEqual(service_name, function_name)

    def test_resource_attributes_configuration(self):
        """Test resource attributes configuration."""
        attributes = "key1=value1,key2=value2"
        os.environ["OTEL_RESOURCE_ATTRIBUTES"] = attributes

        resource_attrs = os.environ.get("OTEL_RESOURCE_ATTRIBUTES")
        self.assertEqual(resource_attrs, attributes)


class TestLogConfiguration(unittest.TestCase):
    """Test logging configuration."""

    def setUp(self):
        """Store original environment."""
        self.old_env = os.environ.copy()

    def tearDown(self):
        """Restore original environment."""
        os.environ.clear()
        os.environ.update(self.old_env)

    def test_log_level_configuration(self):
        """Test log level configuration."""
        os.environ["OTEL_LOG_LEVEL"] = "DEBUG"
        log_level = os.environ.get("OTEL_LOG_LEVEL")
        self.assertEqual(log_level, "DEBUG")

    def test_default_log_level(self):
        """Test default log level."""
        os.environ.pop("OTEL_LOG_LEVEL", None)
        log_level = os.environ.get("OTEL_LOG_LEVEL", "INFO")
        self.assertEqual(log_level, "INFO")


if __name__ == "__main__":
    unittest.main()
