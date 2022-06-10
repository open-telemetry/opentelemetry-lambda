import os
import json
import aiohttp
import asyncio
import boto3
import os
import time

from opentelemetry import _metrics
from opentelemetry.exporter.otlp.proto.grpc._metric_exporter import (
    OTLPMetricExporter,
)
import opentelemetry.exporter.otlp.proto.grpc.version

print( "opentelemetry.exporter.otlp.proto.grpc version", opentelemetry.exporter.otlp.proto.grpc.version.__version__)

from opentelemetry._metrics import (
    get_meter_provider,
    set_meter_provider,
)
from opentelemetry.sdk._metrics import MeterProvider
from opentelemetry.sdk._metrics.export import PeriodicExportingMetricReader

exporter = OTLPMetricExporter(insecure=True)
reader = PeriodicExportingMetricReader(exporter)
provider = MeterProvider(metric_readers=[reader])
set_meter_provider(provider)

meter = get_meter_provider().get_meter("otel_stack_function", "0.1.2")


async def fetch(session, url):
    async with session.get(url) as response:
        return await response.text()


async def callAioHttp():
    async with aiohttp.ClientSession() as session:
        html = await fetch(session, "http://httpbin.org/")

s3 = boto3.resource("s3")

# lambda function
def lambda_handler(event, context):

    loop = asyncio.get_event_loop()
    loop.run_until_complete(callAioHttp())
    
    counter = meter.create_counter(name="first_counter", description="TODO", unit="1",)

    i = 0
    for bucket in s3.buckets.all():
        i = i + 1
        # print(bucket.name)
        
    counter.add(i, attributes={"invocation": context.aws_request_id})

    provider.force_flush() # needed to be sure that metrics are sent to OTel before shutting down the worker
    time.sleep(1)  # This is needed because at the moment there is no API to force the collector to push the message before the Worker is suspended
    
    
    return {"statusCode": 200, "body": json.dumps(os.environ.get("_X_AMZN_TRACE_ID"))}
