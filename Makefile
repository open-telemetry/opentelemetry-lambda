build:
	GOOS=linux GOARCH=amd64 go build -o bin/opentelemetry-lambda-extension main.go

publish:
	zip -r extension.zip bin/
	aws lambda publish-layer-version \
		--layer-name "go-example-extension" \
		--region <use your region> \
		--zip-file  "fileb://extension.zip"

test:
	# TODO