build:
	GOOS=linux GOARCH=amd64 go build -o bin/opentelemetry-lambda-extension *.go

publish:
	zip -r extension.zip bin/
	aws lambda publish-layer-version \
		--layer-name "opentelemetry-extension" \
		--region us-east-1 \
		--zip-file  "fileb://extension.zip"

build-OpenTelemetryExtensionLayer:
	GOOS=linux GOARCH=amd64 go build -o $(ARTIFACTS_DIR)/extensions/opentelemetry-lambda-extension *.go
	chmod +x $(ARTIFACTS_DIR)/extensions/opentelemetry-lambda-extension

run-OpenTelemetryExtensionLayer:
	go run opentelemetry-lambda-extension/*.go

test:
	# TODO
