#!/bin/bash

set -euf -o pipefail

rm -rf ./build/workspace/node_modules

# Space separated list of external NPM packages
EXTERNAL_PACKAGES=( "import-in-the-middle" )

for EXTERNAL_PACKAGE in "${EXTERNAL_PACKAGES[@]}"
do
  echo "Installing external package $EXTERNAL_PACKAGE ..."

  PACKAGE_VERSION=$(npm query "#$EXTERNAL_PACKAGE" \
    | grep version \
    | head -1 \
    | awk -F: '{ print $2 }' \
    | sed 's/[",]//g')

  echo "Resolved version of the external package $EXTERNAL_PACKAGE: $PACKAGE_VERSION"

  npm install "$EXTERNAL_PACKAGE@$PACKAGE_VERSION" --prefix ./build/workspace --production --ignore-scripts

  echo "Installed external package $EXTERNAL_PACKAGE"
done
