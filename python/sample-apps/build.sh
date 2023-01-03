#!/bin/sh
set -e

mkdir -p build/python
python3 -m pip install -r function/requirements.txt -t build/python
cp function/lambda_function.py -t build/python
cd build/python
zip -r ../function.zip ./*
