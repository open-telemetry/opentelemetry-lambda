#!/bin/bash

set -e
set -u

echo_usage() {
	echo "usage: Deploy Lambda layer/application by SAM"
	echo " -r <aws region>"
	echo " -t <cloudformation template>"
	echo " -b <sam build>"
	echo " -d <deploy>"
	echo " -n <specify layer name>"
	echo " -l <show layer arn>"
	echo " -s <stack name>"
}

is_sample() {
	if [[ $(pwd) == *"sample"* ]]; then
		echo 1
	else
		echo 0
	fi
}

main() {
	echo "running..."
	saved_args="$@"
	template='template.yml'
	build=false
	deploy=false
	layer=false

	region=${AWS_REGION-$(aws configure get region)}
	stack=${OTEL_LAMBDA_STACK-"otel-stack"}
	layerName=${OTEL_LAMBDA_LAYER-"otel-layer"}


	collectorPath=${COLLECTOR_PATH-"../../collector"}

	while getopts "hbdxlr:t:s:n:" opt; do
		case "${opt}" in
		h)
			echo_usage
			exit 0
			;;
		b)
			build=true
			;;
		d)
			deploy=true
			;;
		n)
			layerName="${OPTARG}"
			;;
		l)
			layer=true
			;;
		r)
			region="${OPTARG}"
			;;
		t)
			template="${OPTARG}"
			;;
		s)
			stack="${OPTARG}"
			;;
		\?)
			echo "Invalid option: -${OPTARG}" >&2
			exit 1
			;;
		:)
			echo "Option -${OPTARG} requires an argument" >&2
			exit 1
			;;
		esac
	done

	echo "Invoked with: ${saved_args}"

	if [[ $build == false && $deploy == false && $layer == false ]]; then
		build=true
		deploy=true
		layer=true
	fi

	if [[ $build == true ]]; then
		echo "sam building..."
		rm -rf .aws-sam
		rm -rf otel/otel_collector
		mkdir -p otel/otel_collector
		cp -r "$collectorPath"/* otel/otel_collector
		sam build -u -t $template
	fi

	if [[ $deploy == true ]]; then
		if [[ $(is_sample) == "1" ]]; then
			echo "sam deploying..."
			sam deploy --stack-name $stack --region $region --capabilities CAPABILITY_NAMED_IAM --resolve-s3
		else
			echo "zip and awscli deploying..."
			BUCKET_NAME="lambda-artifacts"-$(dd if=/dev/random bs=8 count=1 2>/dev/null | od -An -tx1 | tr -d ' \t\n')
			echo "$BUCKET_NAME"
			cd .aws-sam/build/OTelLayer && zip -q -r layer.zip *
			aws s3 mb s3://$BUCKET_NAME --region $region
			aws s3 cp layer.zip s3://$BUCKET_NAME --region $region
			aws lambda publish-layer-version --region $region --layer-name $layerName --content S3Bucket=$BUCKET_NAME,S3Key=layer.zip --compatible-runtimes nodejs12.x nodejs10.x java11 python3.8 python3.7 --query 'LayerVersionArn' --output text
			aws s3 rm s3://$BUCKET_NAME/layer.zip --region $region
			aws s3 rb s3://$BUCKET_NAME --region $region
			rm layer.zip
			cd ../../..
		fi
		rm -rf otel/otel_collector
		rm -rf .aws-sam
	fi

	if [[ $layer == true ]]; then
		echo -e "\nOTel Lambda layer ARN:"
		if [[ $(is_sample) == "1" ]]; then
			echo $(aws cloudformation describe-stack-resources --stack-name $stack --region $region --query 'StackResources[0].PhysicalResourceId' --output text)
		else
			arn=$(aws lambda list-layer-versions --layer-name $layerName --region $region --query 'max_by(LayerVersions, &Version).LayerVersionArn')
			echo $arn | sed 's/\"//g'
		fi
	fi
}

main "$@"
