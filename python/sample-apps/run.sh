#!/bin/bash
set -euf -o pipefail

cp -r ../src/otel .
cp ../src/run.sh layer.sh
./layer.sh "$@"
rm -rf .aws-sam
rm -rf otel
rm -r layer.sh
