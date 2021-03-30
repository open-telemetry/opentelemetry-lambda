import os
import json
import aiohttp
import asyncio
import boto3


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

    for bucket in s3.buckets.all():
        print(bucket.name)

    return {"statusCode": 200, "body": json.dumps(os.environ.get("_X_AMZN_TRACE_ID"))}
