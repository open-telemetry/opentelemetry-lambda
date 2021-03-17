#!/bin/bash

set -e
set -u

cd layerARNsPy38
for file in ./*
    do
        arn=$(cat $file)
        region=${file##*/}
        echo $arn
        version=$(aws lambda get-layer-version-by-arn --arn $arn --region $region --query 'Version')
        let len=${#arn}-${#version}-1
        layerName=${arn:0:$len}
        (aws lambda add-layer-version-permission --region $region --layer-name $layerName --version-number $version --principal "*" --statement-id publish --action lambda:GetLayerVersion) || true
    done
