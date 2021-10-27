import logging
import os

from importlib import import_module
from pkg_resources import iter_entry_points

from opentelemetry.instrumentation.dependencies import get_dist_dependency_conflicts
from opentelemetry.instrumentation.aws_lambda import AwsLambdaInstrumentor
from opentelemetry.environment_variables import OTEL_PYTHON_DISABLED_INSTRUMENTATIONS
from opentelemetry.instrumentation.distro import BaseDistro, DefaultDistro

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

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

def _load_configurators():
    configured = None
    for entry_point in iter_entry_points("opentelemetry_configurator"):
        if configured is not None:
            logger.warning(
                "Configuration of %s not loaded, %s already loaded",
                entry_point.name,
                configured,
            )
            continue
        try:
            entry_point.load()().configure()  # type: ignore
            configured = entry_point.name
        except Exception as exc:  # pylint: disable=broad-except
            logger.debug("Configuration of %s failed", entry_point.name)


def modify_module_name(module_name):
    """Returns a valid modified module to get imported"""
    return ".".join(module_name.split("/"))

class HandlerError(Exception):
    pass

distro = _load_distros()
distro.configure()
_load_configurators()
_load_instrumentors(distro)
# TODO: move to python-contrib
AwsLambdaInstrumentor().instrument(skip_dep_check=True)

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
