import * as cdk from "aws-cdk-lib";
import * as lambda from "aws-cdk-lib/aws-lambda";
import * as logs from "aws-cdk-lib/aws-logs";
import type { Construct } from "constructs";
import { Stack, StackProps } from "aws-cdk-lib";

export interface IntegrationTestStackProps extends StackProps {
  runtime: lambda.Runtime;
  handler: string;
  handlerCodePath: string;
  collectorLayerZipPath: string;
  instrumentationLayerZipPath: string;
}

export class IntegrationTestStack extends Stack {
  constructor(scope: Construct, id: string, props: IntegrationTestStackProps) {
    super(scope, id, props);

    const collectorLayer = new lambda.LayerVersion(this, "CollectorLayer", {
      code: lambda.Code.fromAsset(props.collectorLayerZipPath),
      compatibleArchitectures: [lambda.Architecture.X86_64],
    });

    const instrumentationLayer = new lambda.LayerVersion(this, "InstrumentationLayer", {
      code: lambda.Code.fromAsset(props.instrumentationLayerZipPath),
      compatibleArchitectures: [lambda.Architecture.X86_64],
    });

    const lambdaFunction = new lambda.Function(this, "TestFunction", {
      runtime: props.runtime,
      handler: props.handler,
      code: lambda.Code.fromAsset(props.handlerCodePath),
      layers: [collectorLayer, instrumentationLayer],
      environment: {
        AWS_LAMBDA_EXEC_WRAPPER: "/opt/otel-handler",
      },
      logGroup: new logs.LogGroup(this, "FunctionLogGroup", {
        retention: logs.RetentionDays.ONE_DAY,
        removalPolicy: cdk.RemovalPolicy.DESTROY,
      }),
      timeout: cdk.Duration.seconds(30),
      memorySize: 512,
    });

    new cdk.CfnOutput(this, "FunctionName", { value: lambdaFunction.functionName });
    new cdk.CfnOutput(this, "LogGroupName", { value: lambdaFunction.logGroup.logGroupName });
  }
}
