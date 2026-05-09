import {
  Toolkit,
  NonInteractiveIoHost,
  StackSelectionStrategy,
} from "@aws-cdk/toolkit-lib";
import * as cdk from "aws-cdk-lib";
import * as lambda from "aws-cdk-lib/aws-lambda";
import type { TestProject } from "vitest/node";
import { IntegrationTestStack } from "./cdk/stack.js";
import * as path from "node:path";

declare module "vitest" {
  export interface ProvidedContext {
    functionName: string;
    logGroupName: string;
  }
}

const LANGUAGE_CONFIG: Record<
  string,
  { runtime: lambda.Runtime; handler: string; handlerDir: string }
> = {
  nodejs: {
    runtime: lambda.Runtime.NODEJS_24_X,
    handler: "index.handler",
    handlerDir: "handlers/nodejs",
  },
  python: {
    runtime: lambda.Runtime.PYTHON_3_14,
    handler: "lambda_function.lambda_handler",
    handlerDir: "handlers/python",
  },
};

export async function setup({ provide }: TestProject) {
  const language = process.env.TEST_LANGUAGE;
  const collectorZip = process.env.COLLECTOR_LAYER_ZIP;
  const instrumentationZip = process.env.INSTRUMENTATION_LAYER_ZIP;

  if (!language || !collectorZip || !instrumentationZip) {
    throw new Error(
      "Required env vars: TEST_LANGUAGE, COLLECTOR_LAYER_ZIP, INSTRUMENTATION_LAYER_ZIP",
    );
  }

  const config = LANGUAGE_CONFIG[language];
  if (!config) {
    throw new Error(
      `Unsupported language: ${language}. Supported: ${Object.keys(LANGUAGE_CONFIG).join(", ")}`,
    );
  }

  const stackName = `IntegrationTest-${language}`;

  const toolkit = new Toolkit({
    ioHost: new NonInteractiveIoHost(),
  });

  const source = await toolkit.fromAssemblyBuilder(async (props) => {
    const app = new cdk.App({ outdir: props.outdir, context: props.context });
    new IntegrationTestStack(app, stackName, {
      runtime: config.runtime,
      handler: config.handler,
      handlerCodePath: path.resolve(config.handlerDir),
      collectorLayerZipPath: path.resolve(collectorZip),
      instrumentationLayerZipPath: path.resolve(instrumentationZip),
    });
    return app.synth();
  });

  const result = await toolkit.deploy(source, {
    stacks: {
      strategy: StackSelectionStrategy.ALL_STACKS,
    },
  });

  const stack = result.stacks[0];
  if (!stack) {
    throw new Error(`Deploy of ${stackName} returned no stacks`);
  }
  const { FunctionName, LogGroupName } = stack.outputs;
  if (!FunctionName || !LogGroupName) {
    throw new Error(
      `Stack ${stackName} missing required outputs (got: ${Object.keys(stack.outputs).join(", ")})`,
    );
  }

  provide("functionName", FunctionName);
  provide("logGroupName", LogGroupName);

  return async () => {
    await toolkit.destroy(source, {
      stacks: {
        strategy: StackSelectionStrategy.ALL_STACKS,
      },
    });
  };
}
