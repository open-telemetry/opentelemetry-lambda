BASE_SPACE:=$(shell pwd)
BUILD_SPACE:=$(BASE_SPACE)/build
BUCKET_NAME:=lambda-artifacts-$(shell dd if=/dev/random bs=8 count=1 2>/dev/null | od -An -tx1 | tr -d ' \t\n')
LAYER_NAME:=otel-collector

VERSION=$(shell cat VERSION)
GIT_SHA=$(shell git rev-parse HEAD)
GOBUILD=GO111MODULE=on CGO_ENABLED=0 installsuffix=cgo go build -trimpath
BUILD_INFO_IMPORT_PATH=main

LDFLAGS=-ldflags "-s -w -X $(BUILD_INFO_IMPORT_PATH).GitHash=$(GIT_SHA) -X $(BUILD_INFO_IMPORT_PATH).Version=$(VERSION) \
-X github.com/open-telemetry/opentelemetry-collector-contrib/exporter/awsxrayexporter.collectorDistribution=opentelemetry-collector-lambda"

clean:
	rm -rf build/

build: clean
	@echo Building otel collector extension
	mkdir -p $(BUILD_SPACE)/extensions
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_SPACE)/extensions .

package:
	@echo Package zip file for collector extension layer
	mkdir -p $(BUILD_SPACE)/collector-config
	cp config* $(BUILD_SPACE)/collector-config
	cd $(BUILD_SPACE) && zip -r collector-extension.zip collector-config extensions

publish-layer: build package
	@echo Publishing collector extension layer...
	aws s3 mb s3://$(BUCKET_NAME)
	aws s3 cp $(BUILD_SPACE)/collector-extension.zip s3://$(BUCKET_NAME)
	aws lambda publish-layer-version --layer-name $(LAYER_NAME) --content S3Bucket=$(BUCKET_NAME),S3Key=collector-extension.zip --compatible-runtimes nodejs12.x nodejs10.x java11 python3.8 --query 'LayerVersionArn' --output text
	@echo Clearing cached files...
	aws s3 rm s3://$(BUCKET_NAME)/collector-extension.zip
	aws s3 rb s3://$(BUCKET_NAME)
	@echo OpenTelemetry Collector layer published.
