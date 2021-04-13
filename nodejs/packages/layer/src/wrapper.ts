import { NodeTracerProvider } from '@opentelemetry/node';
import { SimpleSpanProcessor } from '@opentelemetry/tracing';
import { AwsLambdaInstrumentation } from '@opentelemetry/instrumentation-aws-lambda';
import { registerInstrumentations } from '@opentelemetry/instrumentation';
import { CollectorTraceExporter } from '@opentelemetry/exporter-collector-proto';
import { awsLambdaDetector } from '@opentelemetry/resource-detector-aws';
import { AwsInstrumentation } from 'opentelemetry-instrumentation-aws-sdk';

console.log('Registering OpenTelemetry');

const instrumentations = [
  new AwsInstrumentation({
    suppressInternalInstrumentation: true,
  }),
  new AwsLambdaInstrumentation(),
];

// Register instrumentations synchronously to ensure code is patched even before provider is ready.
registerInstrumentations({
  instrumentations,
});

async function initializeProvider() {
  const resource = await awsLambdaDetector.detect();
  const tracerProvider = new NodeTracerProvider({
    resource,
  });
  // TODO(anuraaga): Switch to BatchSpanProcessor after using published instrumentation package.
  tracerProvider.addSpanProcessor(
    new SimpleSpanProcessor(new CollectorTraceExporter())
  );
  tracerProvider.register();

  // Re-register instrumentation with initialized provider. Patched code will see the update.
  registerInstrumentations({
    instrumentations,
    tracerProvider,
  });
}
initializeProvider();
