# OpenTelemetry Lambda Layer Release Procedure

The release process is almost entirely managed by [GitHub actions](https://github.com/open-telemetry/opentelemetry-lambda/tree/main/.github/workflows). To publish a new layer:

1. Create a new tag for the layer to publish. For example, to create a new collector layer, the following command is used:
        `git tag layer-collector/0.0.8`
2. Push the tag to [opentelemetry-lambda](https://github.com/open-telemetry/opentelemetry-lambda) repository to trigger the publish action.
