#!/bin/bash

set -e
set -u

echo_usage () {
    echo "usage: distribution"
}

invoke() {
    apiid=$(aws cloudformation describe-stack-resource --stack-name $stack --region $region --logical-resource-id api --query 'StackResourceDetail.PhysicalResourceId' --output text)
    for ((i=1; i<=$invokes; i ++))
        do
            echo $i
            curl https://$apiid.execute-api.$region.amazonaws.com/api/
            sleep $interval
        done
    echo "invoke complete"
}

main () {
    saved_args="$@"
    invokes=0
    interval=5
    template='template.yml'
    layer=false
    function=false
    deleteResources=false
    endpoint=false
    stack=${STACK-"otel-sample"}
    region=${AWS_REGION-$(aws configure get region)}

    while getopts "elfcr:s:n:i:t:" opt; do
        case "${opt}" in
            h) echo_usage
                exit 0
                ;;
            r) region="${OPTARG}"
                ;;
            s) stack="${OPTARG}"
                ;;
            n) invokes=${OPTARG}
                ;;
            i) interval=${OPTARG}
                ;;
            t) template=${OPTARG}
                ;;
            l) layer=true
                ;;
            f) function=true
                ;;
            c) endpoint=true
                ;;
            e) deleteResources=true
                ;;
            \?) echo "Invalid option: -${OPTARG}" >&2
                exit 1
                ;;
            :)  echo "Option -${OPTARG} requires an argument" >&2
                exit 1
                ;;
        esac
    done

    if [[ $invokes != 0 ]]; then
        invoke
    fi

    if [[ $layer == true ]]; then
        functionName=$(aws cloudformation describe-stack-resource --stack-name $stack --region $region --logical-resource-id function --query 'StackResourceDetail.PhysicalResourceId' --output text)
        layerArn=$(aws lambda get-function --function-name $functionName --region $region --query 'Configuration.Layers[0].Arn' --output text)
        echo $layerArn
    fi

    if [[ $function == true ]]; then
        functionName=$(aws cloudformation describe-stack-resource --stack-name $stack --region $region --logical-resource-id function --query 'StackResourceDetail.PhysicalResourceId' --output text)
        echo $functionName
    fi

    if [[ $deleteResources == true ]]; then
        aws cloudformation delete-stack --stack-name $stack || true
        aws cloudformation wait stack-delete-complete --stack-name $stack || true
    fi

    if [[ $endpoint == true ]]; then
        apiid=$(aws cloudformation describe-stack-resource --stack-name $stack --region $region --logical-resource-id api --query 'StackResourceDetail.PhysicalResourceId' --output text)
        echo https://$apiid.execute-api.$region.amazonaws.com/api/
    fi
}

main "$@"
