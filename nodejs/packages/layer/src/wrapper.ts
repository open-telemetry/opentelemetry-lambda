import { NodeTracerConfig, NodeTracerProvider } from '@opentelemetry/node';
import {
  SDKRegistrationConfig,
  SimpleSpanProcessor,
} from '@opentelemetry/tracing';
import { AwsLambdaInstrumentation } from '@opentelemetry/instrumentation-aws-lambda';
import { registerInstrumentations } from '@opentelemetry/instrumentation';
import { CollectorTraceExporter } from '@opentelemetry/exporter-collector-proto';
import { awsLambdaDetector } from '@opentelemetry/resource-detector-aws';
import { AwsInstrumentation } from 'opentelemetry-instrumentation-aws-sdk';

declare global {
  function configureTracer(defaultConfig: NodeTracerConfig): NodeTracerConfig;
  function configureSdkRegistration(
    defaultSdkRegistration: SDKRegistrationConfig
  ): SDKRegistrationConfig;
}

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

  let config: NodeTracerConfig = {
    resource,
  };
  if (typeof configureTracer === 'function') {
    config = configureTracer(config);
  }

  const tracerProvider = new NodeTracerProvider(config);
  // TODO(anuraaga): Switch to BatchSpanProcessor after using published instrumentation package.
  tracerProvider.addSpanProcessor(
    new SimpleSpanProcessor(new CollectorTraceExporter())
  );

  let sdkRegistrationConfig: SDKRegistrationConfig = {};
  if (typeof configureSdkRegistration === 'function') {
    sdkRegistrationConfig = configureSdkRegistration(sdkRegistrationConfig);
  }
  tracerProvider.register(sdkRegistrationConfig);

  // Re-register instrumentation with initialized provider. Patched code will see the update.
  registerInstrumentations({
    instrumentations,
    tracerProvider,
  });
}
initializeProvider();
