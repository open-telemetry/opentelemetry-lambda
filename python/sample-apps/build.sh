#!/bin/sh

mkdir -p build/python
pip3 install -r function/requirements.txt -t build/python
cp function/lambda_function.py -t build/python
cd build/python || exit
zip -r ../function.zip *
