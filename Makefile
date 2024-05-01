.DEFAULT_GOAL := help
SHELL := /bin/bash
GOBIN  = $(shell go env GOPATH)/bin
USER  := $(shell whoami)
DATE  := $(shell date +%FT%T%z)
ACTUAL_PWD := $(PWD)
LAMBDA_ARCH = amd64 arm64

############################################################
# Go Info
CGO_ENABLED := 0
# These flags disable compiler optimizations so that debuggers work
GCFLAGS := -N -l
GOFLAGS     := GOFLAGS="-mod=mod"  # For modcache deps.

UNAME_S?= $(shell uname -s)
ifeq ("${UNAME_S}", "Darwin")
    GOOS = darwin
else
    GOOS = linux
endif

.PHONY: help
help:  # Print list of Makefile targets
	@for f in $(MAKEFILE_LIST); do grep -E ':  #' $${f} | grep -v 'LIST\|BEGIN' | \
	sort -u | awk 'BEGIN {FS = ":.*?# "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'; done

.PHONY: all
all: clean lint format test build dist  # make all common targets for all utilities

.PHONY: build
build:  # build binary of all utilities
build: go-mod
	@set -e; \
		echo "==> Building executable for collector"; \
		cd ${ACTUAL_PWD}/collector; \
		mkdir -p build; \
		for ARCH in $(LAMBDA_ARCH); do \
			mkdir -p build/$${ARCH}/extensions; \
			${GOFLAGS} CGO_ENABLED=${CGO_ENABLED} GOOS=${GOOS} GOARCH=$${ARCH} \
				go build -trimpath -o build/$${ARCH}/extensions/collector -ldflags="-s -w"; \
			mkdir -p build/$${ARCH}/collector-config; \
			cp config* build/$${ARCH}/collector-config/; \
		done

.PHONY: clean
clean:  # clean temporary files of all utilities
	@rm -rf collector/build dist build .cache test_coverage.html test_coverage.txt cp.out junit-report.xml

.PHONY: deps-go
deps-go:  # Install dependencies for Go development.
	@echo "==> Installing dependencies for Go development."
	@(cd /tmp; \
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.57.2; \
		go install golang.org/x/tools/cmd/goimports@latest; \
		go install github.com/incu6us/goimports-reviser@latest; \
		go install github.com/jstemmer/go-junit-report@v0.9.1; \
	)

.PHONY: dist
dist:  # build zip file (for lambda) of all utilities
	@set -e; \
		mkdir -p dist ; \
		echo "==> Preparing Lambda artifacts."; \
		for ARCH in $(LAMBDA_ARCH); do \
			(cd  ${ACTUAL_PWD}/collector/build/$${ARCH}; \
				zip -9 ${ACTUAL_PWD}/dist/collector-$${ARCH}.zip collector-config extensions); \
		done

.PHONY: format
format:  # Format go files
	@echo "==> do not format code to avoid unnecesary changes"

.PHONY: lint
lint:  # Lint all Go source code using GolangCI-Lint.
	@set -e; \
		cd ${ACTUAL_PWD}/collector; \
		echo "==> Linting all Go source code using GolangCI-Lint."; \
		$(GOBIN)/golangci-lint run \
			--modules-download-mode mod \
			--exclude-dirs '(vendor|sample_data|.submodules|.cache|.git)' \
			--timeout 600s

.PHONY: test
test:  # Run tests for all Lambdas
	@set -e; \
		cd ${ACTUAL_PWD}/collector; \
		go test -count 1 -race -coverprofile=../cp.out ./...; \
		cat ../cp.out | awk 'BEGIN {cov=0; stat=0;} \
			$$3!="" { cov+=($$3==1?$$2:0); stat+=$$2; } \
    		END {printf("Total coverage: %.2f%% of statements\n", (cov/stat)*100);}'; \
		$(GOFLAGS) go tool cover -html=../cp.out -o ../test_coverage.html; \
		$(GOFLAGS) go tool cover -func=../cp.out 2>&1 | tee test_coverage.txt; \
		$(GOFLAGS) go test -count 1 -race -timeout 180s -v -cover ./... 2>&1 | $(GOBIN)/go-junit-report > ../junit-report.xml

.PHONY: go-mod
go-mod:  # Vendor all Go dependencies.
	@echo "==> Vendoring all Go dependencies."
	@set -e; \
		cd ${ACTUAL_PWD}/collector; \
		go mod tidy; \
		go mod verify; \
		go mod download
