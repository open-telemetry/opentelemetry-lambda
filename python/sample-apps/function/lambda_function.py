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
print(os.environ)

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

    for bucket in s3.buckets.all():
        counter.add(1, attributes={"hello": bucket.name})
        print("CounterAdd")

        print(bucket.name)

    time.sleep(300)
    return {"statusCode": 200, "body": json.dumps(os.environ.get("_X_AMZN_TRACE_ID"))}
