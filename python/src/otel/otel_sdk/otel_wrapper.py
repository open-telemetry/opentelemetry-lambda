import logging
import os

from importlib import import_module
from pkg_resources import iter_entry_points

from opentelemetry.instrumentation.dependencies import get_dist_dependency_conflicts
from opentelemetry import trace

from opentelemetry.instrumentation.aws_lambda import AwsLambdaInstrumentor
from opentelemetry.environment_variables import OTEL_PYTHON_DISABLED_INSTRUMENTATIONS
from opentelemetry.instrumentation.distro import BaseDistro, DefaultDistro

from opentelemetry.sdk.trace import TracerProvider
from opentelemetry.sdk.trace.export import BatchSpanProcessor
from opentelemetry.sdk.resources import Resource, get_aggregated_resources

# NOTE: These are private methods, and _very_ subject to change. We use them
# because we have to, explained in `context_to_otel_sdk_tracer_provider` below.
from opentelemetry.sdk._configuration import (
    _get_exporter_names,
    _get_id_generator,
    _import_exporters,
    _import_id_generator,
)

from opentelemetry.semconv.resource import ResourceAttributes


logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

from opentelemetry.sdk.extension.aws.resource import AwsLambdaResourceDetector
from opentelemetry.sdk.resources import Resource

# TODO: waiting OTel Python supports env variable config for resource detector
# from opentelemetry.resource import AwsLambdaResourceDetector
# from opentelemetry.sdk.resources import Resource
# resource = Resource.create().merge(AwsLambdaResourceDetector().detect())
# trace.get_tracer_provider.resource = resource

def _load_distros() -> BaseDistro:
    for entry_point in iter_entry_points("opentelemetry_distro"):
        try:
            distro = entry_point.load()()
            if not isinstance(distro, BaseDistro):
                logger.debug(
                    "%s is not an OpenTelemetry Distro. Skipping",
                    entry_point.name,
                )
                continue
            logger.debug(
                "Distribution %s will be configured", entry_point.name
            )
            return distro
        except Exception as exc:  # pylint: disable=broad-except
            logger.debug("Distribution %s configuration failed", entry_point.name)
    return DefaultDistro()


def _load_instrumentors(distro):
    package_to_exclude = os.environ.get(OTEL_PYTHON_DISABLED_INSTRUMENTATIONS, [])
    if isinstance(package_to_exclude, str):
        package_to_exclude = package_to_exclude.split(",")
        # to handle users entering "requests , flask" or "requests, flask" with spaces
        package_to_exclude = [x.strip() for x in package_to_exclude]

    for entry_point in iter_entry_points("opentelemetry_instrumentor"):
        if entry_point.name in package_to_exclude:
            logger.debug(
                "Instrumentation skipped for library %s", entry_point.name
            )
            continue

        try:
            conflict = get_dist_dependency_conflicts(entry_point.dist)
            if conflict:
                logger.debug(
                    "Skipping instrumentation %s: %s",
                    entry_point.name,
                    conflict,
                )
                continue

            # tell instrumentation to not run dep checks again as we already did it above
            distro.load_instrumentor(entry_point, skip_dep_check=True)
            logger.info("Instrumented %s", entry_point.name)
        except Exception as exc:  # pylint: disable=broad-except
            logger.debug("Instrumenting of %s failed", entry_point.name)


def modify_module_name(module_name):
    """Returns a valid modified module to get imported"""
    return ".".join(module_name.split("/"))

class HandlerError(Exception):
    pass

distro = _load_distros()
distro.configure()


def context_to_otel_sdk_tracer_provider(lambda_context):
    """Sets and gets the global TracerProvider using the OpenTelemetry Python
    SDK Implemention using the Lambda Context.

    NOTE: We would have liked to let the `opentelemetry_configurator` default
    class `from opentelemetry.sdk._configuration` do this for us but we cannot.

    We _must_ wait to set the `TracerProvider` until the `lambda_context` is
    available because we want to add an attribute to the OpenTelemetry Python
    SDK `Resource` on the `TracerProvider. There is no other opportunity to set
    this because the `TracerProvider` and `Resource` are immutable.

    Args:
        lambda_context: defined by the AWS Lambda service, contains metadata
            like the `invoked_function_arn`.
    Returns:
        The `TracerProvider` it just created and set using the `lambda_context`
    """
    id_generator_name = _get_id_generator()
    id_generator = _import_id_generator(id_generator_name)

    tracer_provider = TracerProvider(
        id_generator=id_generator(),
        resource=Resource(
            {ResourceAttributes.FAAS_ID: lambda_context.invoked_function_arn}
        ).merge(get_aggregated_resources([AwsLambdaResourceDetector()])),
    )

    exporter_names = _get_exporter_names()
    trace_exporters = _import_exporters(exporter_names)

    for _, exporter_class in trace_exporters.items():
        exporter_args = {}
        tracer_provider.add_span_processor(
            BatchSpanProcessor(exporter_class(**exporter_args))
        )

    trace.set_tracer_provider(tracer_provider)

    return trace.get_tracer_provider()


AwsLambdaInstrumentor().instrument(
    context_to_trace_provider=context_to_otel_sdk_tracer_provider,
)
_load_instrumentors(distro)

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
