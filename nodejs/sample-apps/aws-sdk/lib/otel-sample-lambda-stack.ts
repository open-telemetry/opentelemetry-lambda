import * as cdk from 'aws-cdk-lib';
import { Construct } from 'constructs';
import * as lambda from 'aws-cdk-lib/aws-lambda';
import * as path from 'path';

const AWS_ACCOUNT_ID = '184161586896'; // Replace with your AWS account ID if you want to use a specific layer
const NODEJS_LAYER_VERSION = '0_14_0'; // Update with the latest version if needed
const COLLECTOR_LAYER_VERSION = '0_15_0'; // Update with the latest version if needed

export class OtelSampleLambdaStack extends cdk.Stack {
  constructor(scope: Construct, id: string, props?: cdk.StackProps) {
    super(scope, id, props);

    const region = cdk.Stack.of(this).region;
    const architecture = lambda.Architecture.ARM_64;
    const collectorLayerArn = `arn:aws:lambda:${region}:${AWS_ACCOUNT_ID}:layer:opentelemetry-collector-${architecture}-${COLLECTOR_LAYER_VERSION}:1`;
    const nodejsInstrumentationLayerArn = `arn:aws:lambda:${region}:${AWS_ACCOUNT_ID}:layer:opentelemetry-nodejs-${NODEJS_LAYER_VERSION}:1`;

    new lambda.Function(this, 'MyLambdaFunction', {
      runtime: lambda.Runtime.NODEJS_22_X,
      handler: 'index.handler',
      code: lambda.Code.fromAsset(path.join(__dirname, '../build/lambda')),
      layers: [
        lambda.LayerVersion.fromLayerVersionArn(this, 'OtelCollectorLayer', collectorLayerArn),
        lambda.LayerVersion.fromLayerVersionArn(this, 'NodeJsInstrumentationLayer', nodejsInstrumentationLayerArn)
      ],
      timeout: cdk.Duration.seconds(30),
      memorySize: 256,
      architecture,
      environment: {
        OTEL_EXPORTER_OTLP_ENDPOINT: 'http://localhost:4318/',
        OTEL_TRACES_EXPORTER: 'console',
        OTEL_METRICS_EXPORTER: 'console',
        OTEL_LOG_LEVEL: 'DEBUG',
        OTEL_TRACES_SAMPLER: 'always_on',
        AWS_LAMBDA_EXEC_WRAPPER: '/opt/otel-handler',
      },
    });
  }
}