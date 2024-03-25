#!/bin/bash

set -euf -o pipefail

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
	saved_args="$*"
	template='template.yml'
	build=false
	deploy=false
	layer=false

	stack=${OTEL_LAMBDA_STACK-"otel-stack"}
	layerName=${OTEL_LAMBDA_LAYER-"otel-layer"}

	while getopts "hbdlr:t:s:n:" opt; do
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

	if [[ $deploy == true && $region == "" ]]; then
		region=${AWS_REGION-$(aws configure get region)}
	fi

	echo "Invoked with: ${saved_args}"

	if [[ $build == false && $deploy == false && $layer == false ]]; then
		build=true
		deploy=true
		layer=true
	fi

	if [[ $build == true ]]; then
		echo "sam building..."

		echo "run.sh: Starting sam build."
		sam build -u -t "$template"
		zip -qr "$layerName".zip .aws-sam/build
	fi

	if [[ $deploy == true ]]; then
		sam deploy --stack-name "$stack" --region "$region" --capabilities CAPABILITY_NAMED_IAM --resolve-s3 --parameter-overrides LayerName="$layerName"
		rm -rf otel/otel_collector
		rm -f "$layerName".zip
	fi

	if [[ $layer == true ]]; then
		echo -e "\nOTel Lambda layer ARN:"
		arn=$(aws lambda list-layer-versions --layer-name "$layerName" --region "$region" --query 'max_by(LayerVersions, &Version).LayerVersionArn')
		echo "${arn//\"/}"
	fi
}

rm -rf .aws-sam

main "$@"
