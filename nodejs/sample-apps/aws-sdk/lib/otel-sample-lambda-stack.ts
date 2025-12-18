import * as cdk from 'aws-cdk-lib';
import { Construct } from 'constructs';
import * as lambda from 'aws-cdk-lib/aws-lambda';
import {NodejsFunction} from 'aws-cdk-lib/aws-lambda-nodejs';

export class OtelSampleLambdaStack extends cdk.Stack {
  constructor(scope: Construct, id: string, props?: cdk.StackProps) {
    super(scope, id, props);

    const architecture = lambda.Architecture.ARM_64;
    const target = 'node24';
    new NodejsFunction(this, 'MyLambdaFunction', {
      functionName: 'instrumentation-test-handler',

      runtime: lambda.Runtime.NODEJS_24_X,
      handler: 'index.handler',
      entry: './lambda/index.ts',
      bundling: {
        forceDockerBundling: false,
        minify: true,
        sourceMap: true,
        target,
        tsconfig: 'tsconfig.json',
        commandHooks: {
          beforeBundling: (_: string, outputDir: string) => [
            // the lambda wrapper which is registered via NODE_OPTIONS for OpenTelemetry instrumentation needs to find its way into the bundle
            `npx esbuild --bundle lambda/lambda-wrapper.ts --outfile=${outputDir}/lambda-wrapper.js --minify --platform=node --tsconfig='tsconfig.json' --target=${target}`,
          ],
          beforeInstall: () => [],
          afterBundling: () => [],
        }
      },
      timeout: cdk.Duration.seconds(30),
      memorySize: 256,
      architecture,
      environment: {
        NODE_OPTIONS: '--enable-source-maps --require lambda-wrapper',
      },
    });
  }
}
