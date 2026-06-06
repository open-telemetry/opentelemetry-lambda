import * as lambda from "aws-cdk-lib/aws-lambda";
import type { Construct } from "constructs";
import {
  CfnOutput,
  CliCredentialsStackSynthesizer,
  Duration,
  RemovalPolicy,
  Stack,
  StackProps,
} from "aws-cdk-lib";
import { LogGroup, RetentionDays } from "aws-cdk-lib/aws-logs";

export interface IntegrationTestStackProps extends StackProps {
  runtime: lambda.Runtime;
  handler: string;
  architecture: lambda.Architecture;
  handlerCodePath: string;
  collectorLayerZipPath: string;
  instrumentationLayerZipPath: string;
}

export class IntegrationTestStack extends Stack {
  constructor(scope: Construct, id: string, props: IntegrationTestStackProps) {
    super(scope, id, {
      ...props,
      synthesizer: new CliCredentialsStackSynthesizer(),
    });

    const collectorLayer = new lambda.LayerVersion(this, "CollectorLayer", {
      layerVersionName: `${this.stackName}-CollectorLayer`,
      code: lambda.Code.fromAsset(props.collectorLayerZipPath),
      compatibleArchitectures: [props.architecture],
    });

    const instrumentationLayer = new lambda.LayerVersion(
      this,
      "InstrumentationLayer",
      {
        layerVersionName: `${this.stackName}-InstrumentationLayer`,
        code: lambda.Code.fromAsset(props.instrumentationLayerZipPath),
        compatibleArchitectures: [props.architecture],
      },
    );

    const lambdaFunction = new lambda.Function(this, "TestFunction", {
      runtime: props.runtime,
      handler: props.handler,
      architecture: props.architecture,
      code: lambda.Code.fromAsset(props.handlerCodePath),
      layers: [collectorLayer, instrumentationLayer],
      environment: {
        AWS_LAMBDA_EXEC_WRAPPER: "/opt/otel-handler",
      },
      logGroup: new LogGroup(this, "FunctionLogGroup", {
        retention: RetentionDays.ONE_DAY,
        removalPolicy: RemovalPolicy.DESTROY,
      }),
      timeout: Duration.seconds(30),
      memorySize: 512,
    });

    new CfnOutput(this, "FunctionName", { value: lambdaFunction.functionName });
    new CfnOutput(this, "LogGroupName", {
      value: lambdaFunction.logGroup.logGroupName,
    });
  }
}
