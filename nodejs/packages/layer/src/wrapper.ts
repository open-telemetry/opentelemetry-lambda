import { NodeTracerProvider } from '@opentelemetry/node';
import { SimpleSpanProcessor } from '@opentelemetry/tracing';
import { AwsLambdaInstrumentation } from '@opentelemetry/instrumentation-aws-lambda';
import { registerInstrumentations } from '@opentelemetry/instrumentation';
import { AwsInstrumentation } from 'opentelemetry-instrumentation-aws-sdk';
import { CollectorTraceExporter } from '@opentelemetry/exporter-collector-proto';

console.log('Registering OpenTelemetry');

const provider = new NodeTracerProvider();
// TODO(anuraaga): Switch to BatchSpanProcessor after using published instrumentation package.
provider.addSpanProcessor(new SimpleSpanProcessor(new CollectorTraceExporter()));

registerInstrumentations({
  instrumentations: [
    new AwsInstrumentation({
      suppressInternalInstrumentation: true,
    }),
    new AwsLambdaInstrumentation(),
  ],
  tracerProvider: provider,
});

provider.register();
