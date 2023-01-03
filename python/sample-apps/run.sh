#!/bin/bash
set -e

cp -r ../src/otel .
cp ../src/run.sh layer.sh
./layer.sh "$@"
rm -rf .aws-sam
rm -rf otel
rm -r layer.sh
