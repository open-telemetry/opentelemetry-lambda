import os
import logging
from opentelemetry import trace
from importlib import import_module

from opentelemetry.sdk.extension.aws.trace import AwsXRayIdsGenerator

from opentelemetry.sdk.resources import Resource
from opentelemetry.sdk.trace import TracerProvider
from opentelemetry.sdk.trace.export import (
    SimpleExportSpanProcessor,
    BatchExportSpanProcessor,
)

from opentelemetry.resource import AwsLambdaResourceDetector

from opentelemetry.instrumentation.aws_lambda import AwsLambdaInstrumentor
from pkg_resources import iter_entry_points

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

# TODO: get_aggregated_resources
resource = Resource.create().merge(AwsLambdaResourceDetector().detect())
trace.set_tracer_provider(
    TracerProvider(
        ids_generator=AwsXRayIdsGenerator(),
        resource=resource,
    )
)

console_exporter = os.environ.get("OTEL_DEBUG", None)
if not console_exporter is None:
    from opentelemetry.sdk.trace.export import ConsoleSpanExporter

    trace.get_tracer_provider().add_span_processor(
        SimpleExportSpanProcessor(ConsoleSpanExporter())
    )
    logger.info("Console exporter initialized.")

ci = os.environ.get("_OTEL_CI", None)
if ci is None:
    from opentelemetry.exporter.otlp.trace_exporter import OTLPSpanExporter

    otlp_exporter = OTLPSpanExporter(endpoint="localhost:55680", insecure=True)
    span_processor = BatchExportSpanProcessor(otlp_exporter)
    trace.get_tracer_provider().add_span_processor(span_processor)
    logger.info("Otlp exporter initialized.")


AwsLambdaInstrumentor().instrument()
# Load instrumentors from entry_points
for entry_point in iter_entry_points("opentelemetry_instrumentor"):
    print(entry_point)
    try:
        entry_point.load()().instrument()  # type: ignore
        logger.debug("Instrumented %s", entry_point.name)

    except Exception:
        logger.debug("Instrumenting of %s failed", entry_point.name)


def modify_module_name(module_name):
    """Returns a valid modified module to get imported"""
    return ".".join(module_name.split("/"))


class HandlerError(Exception):
    pass


path = os.environ.get("ORIG_HANDLER", None)
if path is None:
    raise HandlerError("ORIG_HANDLER is not defined.")
parts = path.rsplit(".", 1)
if len(parts) != 2:
    raise HandlerError("Value %s for ORIG_HANDLER has invalid format." % path)

(mod_name, handler_name) = parts
modified_mod_name = modify_module_name(mod_name)
handler_module = import_module(modified_mod_name)
lambda_handler = getattr(handler_module, handler_name)
