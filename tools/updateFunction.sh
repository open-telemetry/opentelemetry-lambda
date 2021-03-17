#!/bin/bash

set -e
set -u

stack=${STACK-"otel-py38-sample"}
region=${AWS_REGION-$(aws configure get region)}
environment=''
layers=''
appendLayers=''

while getopts "r:s:l:e:a:" opt; do
    case "${opt}" in
        r) region="${OPTARG}"
            ;;
        s) stack="${OPTARG}"
            ;;
        l) layers="${OPTARG}"
            ;;
        e) environment="${OPTARG}"
            ;;
        a) appendLayers="${OPTARG}"
            ;;
        \?) echo "Invalid option: -${OPTARG}" >&2
            exit 1
            ;;
        :)  echo "Option -${OPTARG} requires an argument" >&2
            exit 1
            ;;
    esac
done

function=$(aws cloudformation describe-stack-resource --stack-name $stack --region $region --logical-resource-id function --query 'StackResourceDetail.PhysicalResourceId' --output text)
echo $function

params=''
if [[ -n $layers ]]; then
    params=' --layers '$layers
fi

if [[ -n $appendLayers ]]; then
    params=' --layers '$(./integration.sh -l)' '$appendLayers
fi

if [[ -n $environment ]]; then
    params=$params' --environment '$environment
fi

echo $params

aws lambda update-function-configuration --function-name $function --region $region $params