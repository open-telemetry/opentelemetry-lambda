#!/bin/sh

set -e
set -u

echo_usage() {
	echo "usage: Build Lambda layer/application by SAM"
	echo " -n <specify layer name>"
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
	layerName="otel-ruby-layer"

	while getopts "n:h" opt; do
		case "${opt}" in
		h)
			echo_usage
			exit 0
			;;
		n)
			layerName="${OPTARG}"
			;;
		\?)
			exit 1
			;;
		:)
			echo "Option -${OPTARG} requires an argument" >&2
			exit 1
			;;
		esac
	done

	cd .aws-sam/build/OTelLayer/
	zip -qr ../../../"$layerName".zip ruby/
	cd -
}

main "$@"
