#!/bin/bash

set -e
set -u

region=${AWS_REGION-$(aws configure get region)}
stacks=$(aws cloudformation list-stacks --region $region --stack-status-filter CREATE_COMPLETE --query 'StackSummaries[*].StackName' --output text)

echo $stacks
for stack in $stacks
do
    echo $stack
    if [[ $stack == $1 ]]; then
        echo "clean stack ..."
        $(aws cloudformation delete-stack --stack-name $stack --region $region)
    fi
done
