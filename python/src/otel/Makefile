export AOC=$(shell pwd)/otel_collector
export SDK=$(shell pwd)/otel_sdk

VERSION=$(shell cat $(AOC)/VERSION)
GIT_SHA=$(shell git rev-parse HEAD)
GOBUILD=GO111MODULE=on CGO_ENABLED=0 installsuffix=cgo go build -trimpath
BUILD_INFO_IMPORT_PATH=main

LDFLAGS=-ldflags "-s -w -X $(BUILD_INFO_IMPORT_PATH).GitHash=$(GIT_SHA) -X $(BUILD_INFO_IMPORT_PATH).Version=$(VERSION) \
-X github.com/open-telemetry/opentelemetry-collector-contrib/exporter/awsxrayexporter.collectorDistribution=aws-otel-collector-lambda"

build-OTelLayer:
	mkdir -p $(ARTIFACTS_DIR)/python
	python3 -m pip install -r $(SDK)/requirements.txt -t $(ARTIFACTS_DIR)/python
	python3 -m pip install -r $(SDK)/requirements-nodeps.txt -t $(ARTIFACTS_DIR)/tmp --no-deps
	cp -r $(ARTIFACTS_DIR)/tmp/* $(ARTIFACTS_DIR)/python/
	rm -rf $(ARTIFACTS_DIR)/tmp
	cp -r $(SDK)/* $(ARTIFACTS_DIR)/python
	chmod 755 $(ARTIFACTS_DIR)/python/otel-instrument
	rm -rf $(ARTIFACTS_DIR)/python/boto*
	rm -rf $(ARTIFACTS_DIR)/python/urllib3*

	rpm --rebuilddb
	yum install -y go
	mkdir -p $(ARTIFACTS_DIR)/extensions
	cd $(AOC) && $(GOBUILD) $(LDFLAGS) -o $(ARTIFACTS_DIR)/extensions/collector .
	mkdir -p $(ARTIFACTS_DIR)/collector-config
	cp $(AOC)/config* $(ARTIFACTS_DIR)/collector-config/
