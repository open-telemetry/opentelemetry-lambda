#!/bin/bash

set -e
set -u

main () {
    saved_args="$@"
    # layerName can be layer name or layer arn without version
    layerName=''
    layerArn=false
    layerVersion=false
    region=${AWS_REGION-$(aws configure get region)}

    while getopts "avn:r:" opt; do
        case "${opt}" in
            h) echo_usage
                exit 0
                ;;
            r) region="${OPTARG}"
                ;;
            n) layerName="${OPTARG}"
                ;;
            a) layerArn=true
                ;;
            v) layerVersion=true
                ;;
            \?) echo "Invalid option: -${OPTARG}" >&2
                exit 1
                ;;
            :)  echo "Option -${OPTARG} requires an argument" >&2
                exit 1
                ;;
        esac
    done

    if [[ $layerArn == true ]]; then
        # CLI --output text bug
        arn=$(aws lambda list-layer-versions --layer-name $layerName --region $region --query 'max_by(LayerVersions, &Version).LayerVersionArn')
        echo $arn|sed 's/\"//g'
    fi

    if [[ $layerVersion == true ]]; then
        aws lambda list-layer-versions --layer-name $layerName --region $region --query 'max_by(LayerVersions, &Version).Version'
    fi
}

main "$@"

