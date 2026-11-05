# DecoupleAfterBatch Converter

The `DecoupleAfterBatch` converter automatically modifies the collector's configuration for the Lambda distribution. Its purpose is to ensure that a decouple processor is always present after a batch processor in a pipeline, in order to prevent potential data loss due to the Lambda environment being frozen.

## Behavior

The converter scans the collector's configuration and makes the following adjustments:

1. If a pipeline contains a batch processor with no decouple processor defined after it, the converter will automatically add a decouple processor to the end of the pipeline.

2. If a pipeline contains a batch processor with a decouple processor already defined after it or there is no batch processor defined, the converter will not make any changes to the pipeline configuration.
