#!/bin/sh

set -e
set -u

echo_usage() {
	echo "usage: Build Lambda layer/application by SAM"
	echo " -n <specify layer name>"
}


main() {
	echo "running..."
	layerName="opentelemetry-ruby-layer"

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
	echo "Finished"
}

main "$@"
