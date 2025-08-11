include ./Makefile.Common

BASE_SPACE:=$(shell pwd)
BUILD_SPACE:=$(BASE_SPACE)/build
BUCKET_NAME:=lambda-artifacts-$(shell dd if=/dev/random bs=8 count=1 2>/dev/null | od -An -tx1 | tr -d ' \t\n')
LAYER_NAME:=otel-collector
OTELCOL_VERSION=$(shell cat VERSION)
GIT_SHA=$(shell git rev-parse HEAD)
GOARCH ?= amd64
GOBUILD=GO111MODULE=on CGO_ENABLED=0 installsuffix=cgo go build -trimpath$(if ${BUILDTAGS}, -tags "${BUILDTAGS}",)
BUILD_INFO_IMPORT_PATH=main

LDFLAGS=-ldflags "-s -w -X $(BUILD_INFO_IMPORT_PATH).GitHash=$(GIT_SHA) -X $(BUILD_INFO_IMPORT_PATH).Version=$(OTELCOL_VERSION) \
-X github.com/open-telemetry/opentelemetry-collector-contrib/exporter/awsxrayexporter.collectorDistribution=opentelemetry-collector-lambda"

TOOLS_MOD_DIR := ./internal/tools
ALL_GO_MOD_DIRS := $(shell find . -type f -name 'go.mod' -exec dirname {} \; | sort)
OTEL_GO_MOD_DIRS := $(filter-out $(TOOLS_MOD_DIR), $(ALL_GO_MOD_DIRS))
GOMODULES = $(OTEL_GO_MOD_DIRS) $(PWD)

# Define a delegation target for each module
.PHONY: $(GOMODULES)
$(GOMODULES):
	@echo "Running target '$(TARGET)' in module '$@'"
	$(MAKE) -C $@ $(TARGET)

# Triggers each module's delegation target
.PHONY: for-all-target
for-all-target: $(GOMODULES)

.PHONY: clean
clean:
	rm -rf build/

.PHONY: build
build: clean set-otelcol-version
	@echo Building otel collector extension, $(shell go version)
	mkdir -p $(BUILD_SPACE)/extensions
	GOOS=linux GOARCH=$(GOARCH) $(GOBUILD) $(LDFLAGS) -o $(BUILD_SPACE)/extensions .

.PHONY: package
package: build
	@echo Package zip file for collector extension layer
	mkdir -p $(BUILD_SPACE)/collector-config
	cp config* $(BUILD_SPACE)/collector-config
	cd $(BUILD_SPACE) && zip -r opentelemetry-collector-layer-$(GOARCH).zip collector-config extensions

# Variables for combined layer build
LANGUAGE ?= 
INSTRUMENTATION_LAYER_DIR ?= $(BUILD_SPACE)/instrumentation
INSTRUMENTATION_MANAGER = ../utils/instrumentation-layer-manager.sh

.PHONY: package-combined
package-combined: build
	@if [ -z "$(LANGUAGE)" ]; then \
		echo "Error: LANGUAGE parameter is required for combined layer build"; \
		echo "Usage: make package-combined LANGUAGE=<language> [GOARCH=<arch>]"; \
		echo "Supported languages: nodejs, python, java, dotnet, ruby, go"; \
		exit 1; \
	fi
	@echo Building combined extension layer for $(LANGUAGE)
	mkdir -p $(BUILD_SPACE)/collector-config
	cp config* $(BUILD_SPACE)/collector-config
	@# Check if instrumentation layer is available for this language
	@if $(INSTRUMENTATION_MANAGER) check $(LANGUAGE); then \
		echo "Downloading instrumentation layer for $(LANGUAGE)..."; \
		RESULT=$$($(INSTRUMENTATION_MANAGER) download $(LANGUAGE) $(BUILD_SPACE)/temp $(GOARCH) 2>&1) || { \
			echo "Warning: Could not download instrumentation layer for $(LANGUAGE): $$RESULT"; \
			echo "Building collector-only layer..."; \
		}; \
		if [ -d "$(BUILD_SPACE)/temp/instrumentation" ]; then \
			echo "Including instrumentation layer in combined build..."; \
			cp -r $(BUILD_SPACE)/temp/instrumentation/* $(BUILD_SPACE)/; \
			echo "$$RESULT" | grep "Release tag:" > $(BUILD_SPACE)/instrumentation-version.txt || echo "unknown" > $(BUILD_SPACE)/instrumentation-version.txt; \
			rm -rf $(BUILD_SPACE)/temp; \
		fi; \
	else \
		echo "No instrumentation layer available for $(LANGUAGE), building collector-only layer"; \
	fi
	@# Create combined layer zip
	cd $(BUILD_SPACE) && zip -r otel-$(LANGUAGE)-extension-$(GOARCH).zip collector-config extensions $(shell [ -f "$(BUILD_SPACE)/instrumentation-version.txt" ] && find . -mindepth 1 -not -path "./collector-config*" -not -path "./extensions*" -not -name "*-$(GOARCH).zip" || echo "")
	@echo Combined layer created: otel-$(LANGUAGE)-extension-$(GOARCH).zip

.PHONY: publish
publish:
	aws lambda publish-layer-version --layer-name $(LAYER_NAME) --zip-file fileb://$(BUILD_SPACE)/opentelemetry-collector-layer-$(GOARCH).zip --compatible-runtimes nodejs16.x nodejs18.x nodejs20.x nodejs22.x java11 java17 java21 python3.9 python3.10 python3.11 python3.12 python3.13 --query 'LayerVersionArn' --output text

.PHONY: publish-layer
publish-layer: package
	@echo Publishing collector extension layer...
	aws lambda publish-layer-version --layer-name $(LAYER_NAME) --zip-file fileb://$(BUILD_SPACE)/opentelemetry-collector-layer-$(GOARCH).zip --compatible-runtimes nodejs16.x nodejs18.x nodejs20.x nodejs22.x java11 java17 java21 python3.9 python3.10 python3.11 python3.12 python3.13 --query 'LayerVersionArn' --output text
	@echo OpenTelemetry Collector layer published.

.PHONY: publish-combined
publish-combined: package-combined
	@if [ -z "$(LANGUAGE)" ]; then \
		echo "Error: LANGUAGE parameter is required for combined layer publish"; \
		echo "Usage: make publish-combined LANGUAGE=<language> [GOARCH=<arch>]"; \
		exit 1; \
	fi
	@echo Publishing combined extension layer for $(LANGUAGE)...
	@# Determine compatible runtimes based on language
	@case "$(LANGUAGE)" in \
		nodejs) RUNTIMES="nodejs18.x nodejs20.x nodejs22.x" ;; \
		python) RUNTIMES="python3.9 python3.10 python3.11 python3.12 python3.13" ;; \
		java) RUNTIMES="java11 java17 java21" ;; \
		dotnet) RUNTIMES="dotnet6 dotnet8" ;; \
		ruby) RUNTIMES="ruby3.2 ruby3.3" ;; \
		go) RUNTIMES="provided provided.al2" ;; \
		*) echo "Unknown language: $(LANGUAGE)"; exit 1 ;; \
	esac; \
	aws lambda publish-layer-version \
		--layer-name otel-$(LANGUAGE)-extension$(if $(findstring arm64,$(GOARCH)),-arm64,$(if $(findstring amd64,$(GOARCH)),-amd64,)) \
		--zip-file fileb://$(BUILD_SPACE)/otel-$(LANGUAGE)-extension-$(GOARCH).zip \
		--compatible-runtimes $$RUNTIMES \
		--query 'LayerVersionArn' --output text
	@echo Combined extension layer for $(LANGUAGE) published.

.PHONY: set-otelcol-version
set-otelcol-version:
	@OTELCOL_VERSION=$$(grep "go.opentelemetry.io/collector/otelcol v" go.mod | awk '{print $$2; exit}'); \
	echo $$OTELCOL_VERSION > VERSION; \

.PHONY: gotidy
gotidy:
	@$(MAKE) for-all-target TARGET="tidy"

.PHONY: gotest
gotest:
	@$(MAKE) for-all-target TARGET="test"

.PHONY: goporto
goporto: install-tools
	porto -w --include-internal ./

.PHONY: golint
golint: install-tools
	@$(MAKE) for-all-target TARGET="lint"

.PHONY: goimpi
goimpi: install-tools
	@$(MAKE) for-all-target TARGET="impi"

.PHONY: gofmt
gofmt:
	@$(MAKE) for-all-target TARGET="fmt"

.PHONY: install-tools
install-tools:
	cd $(TOOLS_MOD_DIR) && $(GOCMD) install github.com/client9/misspell/cmd/misspell
	cd $(TOOLS_MOD_DIR) && $(GOCMD) install github.com/golangci/golangci-lint/cmd/golangci-lint
	cd $(TOOLS_MOD_DIR) && $(GOCMD) install github.com/google/addlicense
	cd $(TOOLS_MOD_DIR) && $(GOCMD) install github.com/ory/go-acc
	cd $(TOOLS_MOD_DIR) && $(GOCMD) install github.com/pavius/impi/cmd/impi
	cd $(TOOLS_MOD_DIR) && $(GOCMD) install github.com/tcnksm/ghr
	cd $(TOOLS_MOD_DIR) && $(GOCMD) install github.com/wadey/gocovmerge
	cd $(TOOLS_MOD_DIR) && $(GOCMD) install go.opentelemetry.io/build-tools/chloggen
	cd $(TOOLS_MOD_DIR) && $(GOCMD) install go.opentelemetry.io/build-tools/semconvgen
	cd $(TOOLS_MOD_DIR) && $(GOCMD) install golang.org/x/exp/cmd/apidiff
	cd $(TOOLS_MOD_DIR) && $(GOCMD) install golang.org/x/tools/cmd/goimports
	cd $(TOOLS_MOD_DIR) && $(GOCMD) install github.com/jcchavezs/porto/cmd/porto
	cd $(TOOLS_MOD_DIR) && $(GOCMD) install go.opentelemetry.io/build-tools/multimod
	cd $(TOOLS_MOD_DIR) && $(GOCMD) install go.opentelemetry.io/build-tools/crosslink
