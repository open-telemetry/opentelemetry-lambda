import { NodeTracerProvider } from '@opentelemetry/node';
import { SimpleSpanProcessor } from '@opentelemetry/tracing';
import { AwsLambdaInstrumentation } from '@opentelemetry/instrumentation-aws-lambda';
import { registerInstrumentations } from '@opentelemetry/instrumentation';
import { CollectorTraceExporter } from '@opentelemetry/exporter-collector-proto';
import { CLOUD_RESOURCE, Resource } from '@opentelemetry/resources';
import { AwsInstrumentation } from 'opentelemetry-instrumentation-aws-sdk';

console.log('Registering OpenTelemetry');

// TODO(anuraaga): Replace with detector after figuring out how to deal with async.
// https://github.com/open-telemetry/opentelemetry-js/pull/2102/files#r611385113
const resource = new Resource({
  [CLOUD_RESOURCE.PROVIDER]: 'aws',
  [CLOUD_RESOURCE.REGION]: process.env.AWS_REGION!,
  'faas.name': process.env.AWS_LAMBDA_FUNCTION_NAME!,
  'faas.version': process.env.AWS_LAMBDA_FUNCTION_VERSION!,
});

const provider = new NodeTracerProvider({
  resource,
});
// TODO(anuraaga): Switch to BatchSpanProcessor after using published instrumentation package.
provider.addSpanProcessor(
  new SimpleSpanProcessor(new CollectorTraceExporter())
);

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
