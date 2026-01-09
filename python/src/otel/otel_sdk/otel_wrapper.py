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
`otel_wrapper.py`

This file serves as a wrapper over the user's Lambda function.
"""

import importlib
import logging
import os

from opentelemetry import metrics, trace
from opentelemetry.instrumentation.aws_lambda import AwsLambdaInstrumentor
from opentelemetry.sdk.metrics import MeterProvider
from opentelemetry.sdk.metrics.export import (
    ConsoleMetricExporter,
    PeriodicExportingMetricReader,
)
from opentelemetry.sdk.resources import Resource
from opentelemetry.sdk.trace import TracerProvider
from opentelemetry.sdk.trace.export import (
    BatchSpanProcessor,
    ConsoleSpanExporter,
    SimpleSpanProcessor,
)

# Try to import AWS Lambda resource detector
try:
    from opentelemetry.sdk.extension.aws.resource._lambda import (
        AwsLambdaResourceDetector,
    )
except ImportError:
    AwsLambdaResourceDetector = None

# Environment Variables
OTEL_LOG_LEVEL = "OTEL_LOG_LEVEL"
OTEL_PROPAGATORS = "OTEL_PROPAGATORS"
OTEL_TRACES_EXPORTER = "OTEL_TRACES_EXPORTER"
OTEL_METRICS_EXPORTER = "OTEL_METRICS_EXPORTER"
OTEL_SERVICE_NAME = "OTEL_SERVICE_NAME"
OTEL_RESOURCE_ATTRIBUTES = "OTEL_RESOURCE_ATTRIBUTES"

# Import exporters
try:
    from opentelemetry.exporter.otlp.proto.http.trace_exporter import OTLPSpanExporter
except ImportError:
    OTLPSpanExporter = None

try:
    from opentelemetry.exporter.otlp.proto.http.metric_exporter import (
        OTLPMetricExporter,
    )
except ImportError:
    OTLPMetricExporter = None

logger = logging.getLogger(__name__)

# Default instrumentations for Lambda
# Following the Node.js layer pattern, only core network instrumentations are defaults
# dns, http, net in Node.js -> no exact Python equivalent (http is at framework level)
# AWS SDK is always loaded (not part of defaults in Node.js either, it's added separately)
DEFAULT_INSTRUMENTATIONS = []


def _get_active_instrumentations():
    """Determine which instrumentations should be loaded for Lambda.

    Respects OTEL_PYTHON_ENABLED_INSTRUMENTATIONS and OTEL_PYTHON_DISABLED_INSTRUMENTATIONS.

    Note: Unlike explicit defaults, botocore and aws-lambda instrumentations are
    always loaded as they are essential for Lambda operation (similar to how Node.js
    layer always loads AwsInstrumentation and AwsLambdaInstrumentation).
    """
    enabled = os.environ.get("OTEL_PYTHON_ENABLED_INSTRUMENTATIONS")
    # If explicitly enabled, only load those; otherwise use defaults
    active = set(enabled.split(",")) if enabled else set(DEFAULT_INSTRUMENTATIONS)

    # Remove any disabled instrumentations
    disabled = os.environ.get("OTEL_PYTHON_DISABLED_INSTRUMENTATIONS")
    if disabled:
        for item in disabled.split(","):
            active.discard(item.strip())

    return active


def _load_instrumentations():
    """Load and configure instrumentations for Lambda functions.

    Similar to Node.js createInstrumentations() pattern:
    - AwsInstrumentation (botocore) is always loaded
    - AwsLambdaInstrumentation is always loaded
    - Additional instrumentations can be enabled via environment variables

    Conditionally loads instrumentations based on environment variables:
    - OTEL_PYTHON_ENABLED_INSTRUMENTATIONS: Only load specified instrumentations (comma-separated)
    - OTEL_PYTHON_DISABLED_INSTRUMENTATIONS: Disable specific instrumentations (comma-separated)

    Available optional instrumentations:
    - HTTP clients: requests, aiohttp-client, urllib, urllib3
    - Web frameworks: django, flask, fastapi, starlette, falcon, pyramid, tornado
    - Databases: psycopg2, pymongo, pymysql, mysql, asyncpg, sqlite3, sqlalchemy
    - AWS services: boto, boto3sqs
    - Messaging: celery, redis
    - Other: grpc, jinja2, pymemcache, elasticsearch, wsgi, asgi, dbapi

    Example: OTEL_PYTHON_ENABLED_INSTRUMENTATIONS=requests,psycopg2,redis
    """
    active_instrumentations = _get_active_instrumentations()

    # Instrumentation registry - maps names to import paths and instrumentor classes
    # Format: "name": ("module.path", "InstrumentorClass")
    INSTRUMENTATIONS = {
        # botocore is always loaded (see below), included here for completeness
        "botocore": (
            "opentelemetry.instrumentation.botocore",
            "BotocoreInstrumentor",
        ),
        # HTTP Clients
        "requests": (
            "opentelemetry.instrumentation.requests",
            "RequestsInstrumentor",
        ),
        "aiohttp-client": (
            "opentelemetry.instrumentation.aiohttp_client",
            "AioHttpClientInstrumentor",
        ),
        "urllib": ("opentelemetry.instrumentation.urllib", "URLLibInstrumentor"),
        "urllib3": ("opentelemetry.instrumentation.urllib3", "URLLib3Instrumentor"),
        # Web Frameworks
        "django": ("opentelemetry.instrumentation.django", "DjangoInstrumentor"),
        "flask": ("opentelemetry.instrumentation.flask", "FlaskInstrumentor"),
        "fastapi": ("opentelemetry.instrumentation.fastapi", "FastAPIInstrumentor"),
        "starlette": (
            "opentelemetry.instrumentation.starlette",
            "StarletteInstrumentor",
        ),
        "falcon": ("opentelemetry.instrumentation.falcon", "FalconInstrumentor"),
        "pyramid": ("opentelemetry.instrumentation.pyramid", "PyramidInstrumentor"),
        "tornado": ("opentelemetry.instrumentation.tornado", "TornadoInstrumentor"),
        # Databases
        "psycopg2": (
            "opentelemetry.instrumentation.psycopg2",
            "Psycopg2Instrumentor",
        ),
        "pymongo": ("opentelemetry.instrumentation.pymongo", "PymongoInstrumentor"),
        "pymysql": ("opentelemetry.instrumentation.pymysql", "PyMySQLInstrumentor"),
        "mysql": (
            "opentelemetry.instrumentation.mysql",
            "MySQLInstrumentor",
        ),
        "asyncpg": ("opentelemetry.instrumentation.asyncpg", "AsyncPGInstrumentor"),
        "sqlite3": ("opentelemetry.instrumentation.sqlite3", "SQLite3Instrumentor"),
        "sqlalchemy": (
            "opentelemetry.instrumentation.sqlalchemy",
            "SQLAlchemyInstrumentor",
        ),
        # AWS Services
        "boto": ("opentelemetry.instrumentation.boto", "BotoInstrumentor"),
        "boto3sqs": ("opentelemetry.instrumentation.boto3sqs", "Boto3SQSInstrumentor"),
        # Messaging & Caching
        "celery": ("opentelemetry.instrumentation.celery", "CeleryInstrumentor"),
        "redis": ("opentelemetry.instrumentation.redis", "RedisInstrumentor"),
        "pymemcache": (
            "opentelemetry.instrumentation.pymemcache",
            "PymemcacheInstrumentor",
        ),
        # Search
        "elasticsearch": (
            "opentelemetry.instrumentation.elasticsearch",
            "ElasticsearchInstrumentor",
        ),
        # RPC
        "grpc": ("opentelemetry.instrumentation.grpc", "GrpcInstrumentorClient"),
        # Templating
        "jinja2": ("opentelemetry.instrumentation.jinja2", "Jinja2Instrumentor"),
        # WSGI/ASGI
        "wsgi": ("opentelemetry.instrumentation.wsgi", "OpenTelemetryMiddleware"),
        "asgi": ("opentelemetry.instrumentation.asgi", "OpenTelemetryMiddleware"),
        # Database API
        "dbapi": ("opentelemetry.instrumentation.dbapi", "trace_integration"),
    }

    # botocore (AWS SDK) - ALWAYS loaded for Lambda (like Node.js AwsInstrumentation)
    try:
        from opentelemetry.instrumentation.botocore import BotocoreInstrumentor

        BotocoreInstrumentor().instrument()
        logger.debug("Loaded botocore instrumentation (always enabled)")
    except ImportError:
        logger.warning("botocore instrumentation not available")
    except Exception as e:
        logger.warning(f"Failed to load botocore instrumentation: {e}")

    # Load optional instrumentations based on active_instrumentations
    for name, (module_path, class_name) in INSTRUMENTATIONS.items():
        # Skip botocore since it's always loaded above
        if name == "botocore":
            continue

        # Check if this instrumentation should be loaded
        if name not in active_instrumentations:
            continue

        try:
            # Dynamically import the instrumentation module
            module = importlib.import_module(module_path)
            instrumentor_class = getattr(module, class_name)

            # Special handling for certain instrumentations
            if name in ("wsgi", "asgi"):
                # WSGI/ASGI are middleware, not instrumentors - skip auto-loading
                logger.debug(
                    f"Skipping {name} instrumentation (middleware, not auto-instrumentable)"
                )
                continue
            elif name == "dbapi":
                # dbapi is a function, not a class - skip auto-loading
                logger.debug(
                    f"Skipping {name} instrumentation (requires manual integration)"
                )
                continue

            # Instrument the library
            instrumentor_class().instrument()
            logger.debug(f"Loaded {name} instrumentation")

        except ImportError:
            logger.debug(
                f"{name} instrumentation not available (package not installed)"
            )
        except AttributeError as e:
            logger.warning(f"Failed to find {class_name} in {module_path}: {e}")
        except Exception as e:
            logger.warning(f"Failed to load {name} instrumentation: {e}")


def _configure_logger():
    log_level = os.environ.get(OTEL_LOG_LEVEL, "INFO").upper()
    logging.basicConfig(level=log_level)


def _configure_service_name():
    """Set service name to Lambda function name if not already set"""
    if not os.environ.get(OTEL_SERVICE_NAME):
        function_name = os.environ.get("AWS_LAMBDA_FUNCTION_NAME")
        if function_name:
            # Check if OTEL_RESOURCE_ATTRIBUTES already has service.name
            resource_attrs = os.environ.get(OTEL_RESOURCE_ATTRIBUTES, "")
            if "service.name=" not in resource_attrs:
                if resource_attrs:
                    os.environ[OTEL_RESOURCE_ATTRIBUTES] = (
                        f"service.name={function_name},{resource_attrs}"
                    )
                else:
                    os.environ[OTEL_RESOURCE_ATTRIBUTES] = (
                        f"service.name={function_name}"
                    )


def _configure_propagators():
    """Set default propagators for Lambda if not configured.

    Includes X-Ray propagator for AWS service integration.
    """
    if not os.environ.get(OTEL_PROPAGATORS):
        # tracecontext: W3C standard propagation
        # baggage: W3C baggage propagation
        # xray: AWS X-Ray propagation for AWS service integration
        os.environ[OTEL_PROPAGATORS] = "tracecontext,baggage,xray"


def _get_lambda_resource():
    """Create OpenTelemetry resource with Lambda-specific attributes.

    Returns:
        Resource with Lambda function metadata or empty resource if detector unavailable.
    """
    if AwsLambdaResourceDetector:
        from opentelemetry.sdk.resources import get_aggregated_resources

        return get_aggregated_resources([AwsLambdaResourceDetector()])
    return Resource.create()


def _configure_tracer_provider():
    """Configure OpenTelemetry TracerProvider for Lambda.

    Sets up trace export with Lambda resource detection and configured exporters.
    """
    provider = trace.get_tracer_provider()
    is_proxy = isinstance(provider, trace.ProxyTracerProvider)

    if not is_proxy:
        logger.debug("TracerProvider already configured.")
        return

    logger.debug("Configuring TracerProvider for Lambda...")

    resource = _get_lambda_resource()
    provider = TracerProvider(resource=resource)

    exporter_name = os.environ.get(OTEL_TRACES_EXPORTER, "otlp").lower().strip()

    # Handle "none" exporter - no tracing
    if "none" in exporter_name:
        logger.debug(
            "Traces exporter set to 'none', skipping trace export configuration."
        )
        trace.set_tracer_provider(provider)
        return

    # Support multiple exporters
    exporters = []
    for exp_name in exporter_name.split(","):
        exp_name = exp_name.strip()
        if exp_name == "otlp":
            if OTLPSpanExporter:
                exporters.append(OTLPSpanExporter())
            else:
                logger.warning("OTLP Exporter not installed.")
        elif exp_name == "console":
            exporters.append(ConsoleSpanExporter())
        else:
            logger.warning(f"Unknown exporter: {exp_name}")

    # Add span processors based on exporter type
    for exporter in exporters:
        if isinstance(exporter, ConsoleSpanExporter):
            # Use SimpleSpanProcessor for console exporter (immediate export)
            processor = SimpleSpanProcessor(exporter)
        else:
            # Use BatchSpanProcessor for other exporters
            processor = BatchSpanProcessor(exporter)
        provider.add_span_processor(processor)

    trace.set_tracer_provider(provider)


def _configure_meter_provider():
    """Configure OpenTelemetry MeterProvider for Lambda.

    Sets up metrics export with Lambda resource detection and configured exporters.
    """
    provider = metrics.get_meter_provider()
    is_proxy = isinstance(provider, metrics.ProxyMeterProvider)

    if not is_proxy:
        logger.debug("MeterProvider already configured.")
        return

    logger.debug("Configuring MeterProvider for Lambda...")

    resource = _get_lambda_resource()

    exporter_name = os.environ.get(OTEL_METRICS_EXPORTER, "otlp").lower().strip()

    # Handle "none" exporter - no metrics
    if "none" in exporter_name:
        logger.debug(
            "Metrics exporter set to 'none', skipping metrics export configuration."
        )
        return

    readers = []
    for exp_name in exporter_name.split(","):
        exp_name = exp_name.strip()
        if exp_name == "otlp":
            if OTLPMetricExporter:
                exporter = OTLPMetricExporter()
                readers.append(PeriodicExportingMetricReader(exporter))
            else:
                logger.warning("OTLP Metric Exporter not installed.")
        elif exp_name == "console":
            exporter = ConsoleMetricExporter()
            readers.append(PeriodicExportingMetricReader(exporter))
        else:
            logger.warning(f"Unknown metric exporter: {exp_name}")

    if readers:
        provider = MeterProvider(resource=resource, metric_readers=readers)
        metrics.set_meter_provider(provider)


def modify_module_name(module_name):
    """Convert Lambda handler path format to Python module path.

    Converts "/" in handler path to "." for proper Python import.
    Example: "handlers/main" -> "handlers.main"
    """
    return ".".join(module_name.split("/"))


class HandlerError(Exception):
    pass


# Initialize Configuration
_configure_logger()
_configure_service_name()
_configure_propagators()
_configure_tracer_provider()
_configure_meter_provider()

# Load instrumentations - botocore (AWS SDK) is always loaded
_load_instrumentations()

# Instrument Lambda Handler - ALWAYS loaded (like Node.js AwsLambdaInstrumentation)
# This must be called after tracer provider configuration
AwsLambdaInstrumentor().instrument()

path = os.environ.get("ORIG_HANDLER")

if path is None:
    raise HandlerError("ORIG_HANDLER is not defined.")

try:
    (mod_name, handler_name) = path.rsplit(".", 1)
except ValueError as e:
    raise HandlerError(f"Bad path '{path}' for ORIG_HANDLER: {e!s}") from e

modified_mod_name = modify_module_name(mod_name)
handler_module = importlib.import_module(modified_mod_name)
lambda_handler = getattr(handler_module, handler_name)
