import { DiagConsoleLogger, DiagLogLevel, diag } from '@opentelemetry/api';
import { registerInstrumentations } from '@opentelemetry/instrumentation';
import { AwsLambdaInstrumentation } from '@opentelemetry/instrumentation-aws-lambda';
import { AwsInstrumentation } from '@opentelemetry/instrumentation-aws-sdk';
import { resourceFromAttributes } from '@opentelemetry/resources';
import { AlwaysOnSampler, BatchSpanProcessor, ConsoleSpanExporter} from '@opentelemetry/sdk-trace-base';
import { NodeTracerProvider } from '@opentelemetry/sdk-trace-node';
import { ATTR_SERVICE_NAME } from '@opentelemetry/semantic-conventions';

diag.setLogger(new DiagConsoleLogger(), {
  logLevel: DiagLogLevel.DEBUG,
});

const provider = new NodeTracerProvider({
  spanProcessors: [new BatchSpanProcessor(new ConsoleSpanExporter()) ],
  sampler: new AlwaysOnSampler(),
  resource: resourceFromAttributes({
    [ATTR_SERVICE_NAME]: process.env.AWS_LAMBDA_FUNCTION_NAME
  }),
});

provider.register();

registerInstrumentations({
  instrumentations: [
    new AwsInstrumentation({ suppressInternalInstrumentation: true }),
    new AwsLambdaInstrumentation({}),
  ],
});
