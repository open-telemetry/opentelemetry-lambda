import json
import boto3

sts = boto3.client("sts")


def lambda_handler(event, context):
    identity = sts.get_caller_identity()
    return {
        "statusCode": 200,
        "body": json.dumps({"status": "ok", "account": identity["Account"]}),
    }
