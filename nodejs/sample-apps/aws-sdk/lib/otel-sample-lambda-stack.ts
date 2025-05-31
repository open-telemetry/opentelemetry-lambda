import * as cdk from 'aws-cdk-lib';
import { Construct } from 'constructs';
import * as lambda from 'aws-cdk-lib/aws-lambda';
import * as path from 'path';

const AWS_ACCOUNT_ID = '184161586896'; // Replace with your AWS account ID if you want to use a specific layer

export class OtelSampleLambdaStack extends cdk.Stack {
  constructor(scope: Construct, id: string, props?: cdk.StackProps) {
    super(scope, id, props);

    const region = cdk.Stack.of(this).region;
    const architecture = 'amd64'; // or 'arm64' for ARM architecture
    const collectorLayerArn = `arn:aws:lambda:${region}:${AWS_ACCOUNT_ID}:layer:opentelemetry-collector-${architecture}-0_15_0:1`; // Update with the latest version if needed
    const nodejsInstrumentationLayerArn = `arn:aws:lambda:${region}:${AWS_ACCOUNT_ID}:layer:opentelemetry-nodejs-0_14_0:1`; // Update with the latest version if needed

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
      environment: {
        OTEL_EXPORTER_OTLP_ENDPOINT: 'http://localhost:4318/',
        OTEL_TRACES_EXPORTER: 'console',
        OTEL_METRICS_EXPORTER: 'logging',
        OTEL_LOG_LEVEL: 'INFO',
        OTEL_TRACES_SAMPLER: 'always_on',
        AWS_LAMBDA_EXEC_WRAPPER: '/opt/otel-handler',
      },
    });
  }
}