#!/bin/bash

set -e
set -u

main () {
    saved_args="$@"
    fromLayerArn=''
    toLayerName=''

    while getopts "f:t:" opt; do
        case "${opt}" in
            f) fromLayerArn="${OPTARG}"
                ;;
            t) toLayerName="${OPTARG}"
                ;;
            \?) echo "Invalid option: -${OPTARG}" >&2
                exit 1
                ;;
            :)  echo "Option -${OPTARG} requires an argument" >&2
                exit 1
                ;;
        esac
    done

    echo "Invoked with: ${saved_args}"

    URL=$(aws lambda get-layer-version-by-arn --arn $fromLayerArn --query Content.Location --output text)
    curl $URL -o layer.zip
    BUCKET_NAME="lambda-artifacts"-$(dd if=/dev/random bs=8 count=1 2>/dev/null | od -An -tx1 | tr -d ' \t\n')
    aws s3 mb s3://$BUCKET_NAME 
	aws s3 cp layer.zip s3://$BUCKET_NAME
	aws lambda publish-layer-version --layer-name $toLayerName --content S3Bucket=$BUCKET_NAME,S3Key=layer.zip --query 'LayerVersionArn' --output text
	aws s3 rm s3://$BUCKET_NAME/layer.zip
	aws s3 rb s3://$BUCKET_NAME
    rm layer.zip
}

main "$@"
