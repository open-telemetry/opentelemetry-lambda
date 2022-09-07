import os
import json
import boto3
import requests

s3 = boto3.resource("s3")

# lambda function
def lambda_handler(event, context):

    requests.get("http://httpbin.org/")

    for bucket in s3.buckets.all():
        print(bucket.name)

    return {"statusCode": 200, "body": json.dumps(os.environ.get("_X_AMZN_TRACE_ID"))}
