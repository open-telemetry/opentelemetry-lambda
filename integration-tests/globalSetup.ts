import { resolve } from "node:path";
import {
  Toolkit,
  NonInteractiveIoHost,
  StackSelectionStrategy,
} from "@aws-cdk/toolkit-lib";
import { Runtime } from "aws-cdk-lib/aws-lambda";
import type { TestProject } from "vitest/node";
import { IntegrationTestStack } from "./cdk/stack.js";
import { App, Tags } from "aws-cdk-lib";

declare module "vitest" {
  export interface ProvidedContext {
    functionName: string;
    logGroupName: string;
    language: string;
    expectedInstrumentationScopes: string[];
  }
}

type LanguageConfig = {
  runtime: Runtime;
  handler: string;
  handlerDir: string;
  expectedInstrumentationScopes: string[];
};

const LANGUAGE_CONFIG = {
  nodejs: {
    runtime: Runtime.NODEJS_24_X,
    handler: "index.handler",
    handlerDir: "handlers/nodejs",
    expectedInstrumentationScopes: [
      "@opentelemetry/instrumentation-aws-sdk",
      "@opentelemetry/instrumentation-aws-lambda",
    ],
  },
  python: {
    runtime: Runtime.PYTHON_3_14,
    handler: "lambda_function.lambda_handler",
    handlerDir: "handlers/python",
    expectedInstrumentationScopes: [
      "opentelemetry.instrumentation.botocore",
      "opentelemetry.instrumentation.aws_lambda",
    ],
  },
  ruby: {
    runtime: Runtime.RUBY_3_3,
    handler: "lambda_function.handler",
    handlerDir: "handlers/ruby",
    expectedInstrumentationScopes: [
      "OpenTelemetry::Instrumentation::AwsLambda",
    ],
  },
  javaagent: {
    runtime: Runtime.JAVA_21,
    handler:
      "io.opentelemetry.lambda.integrationtests.StsRequestHandler::handleRequest",
    handlerDir: "handlers/java/build/libs/handler-all.jar",
    expectedInstrumentationScopes: [
      "io.opentelemetry.aws-lambda-events-2.2",
      "io.opentelemetry.aws-sdk-2.2",
    ],
  },
  javawrapper: {
    runtime: Runtime.JAVA_21,
    handler:
      "io.opentelemetry.lambda.integrationtests.StsRequestHandler::handleRequest",
    handlerDir: "handlers/java/build/libs/handler-all.jar",
    expectedInstrumentationScopes: [
      "io.opentelemetry.aws-sdk-2.2",
      "io.opentelemetry.aws-lambda-core-1.0",
    ],
  },
} satisfies Record<string, LanguageConfig>;

type SupportedLanguage = keyof typeof LANGUAGE_CONFIG;

function isSupportedLanguage(language: string): language is SupportedLanguage {
  return language in LANGUAGE_CONFIG;
}

export async function setup({ provide }: TestProject) {
  const language = process.env.TEST_LANGUAGE;
  const collectorZip = process.env.COLLECTOR_LAYER_ZIP;
  const instrumentationZip = process.env.INSTRUMENTATION_LAYER_ZIP;

  if (!language || !collectorZip || !instrumentationZip) {
    throw new Error(
      "Required env vars: TEST_LANGUAGE, COLLECTOR_LAYER_ZIP, INSTRUMENTATION_LAYER_ZIP",
    );
  }

  if (!isSupportedLanguage(language)) {
    throw new Error(`Unsupported language: ${language}`);
  }
  const config = LANGUAGE_CONFIG[language];

  const runId = process.env.GITHUB_RUN_ID;
  const runAttempt = process.env.GITHUB_RUN_ATTEMPT;
  const stackName = runId
    ? `IntegrationTest-${language}-${runId}-${runAttempt}`
    : `IntegrationTest-${language}`;

  const toolkit = new Toolkit({
    ioHost: new NonInteractiveIoHost(),
  });

  const source = await toolkit.fromAssemblyBuilder(async (props) => {
    const app = new App({ outdir: props.outdir, context: props.context });

    Tags.of(app).add("Purpose", "integration-test");
    Tags.of(app).add("Language", language);
    if (runId) {
      Tags.of(app).add("GitHubRunId", runId);
      Tags.of(app).add("GitHubRunAttempt", runAttempt ?? "1");
    }

    new IntegrationTestStack(app, stackName, {
      runtime: config.runtime,
      handler: config.handler,
      handlerCodePath: resolve(config.handlerDir),
      collectorLayerZipPath: resolve(collectorZip),
      instrumentationLayerZipPath: resolve(instrumentationZip),
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
  provide("language", language);
  provide(
    "expectedInstrumentationScopes",
    config.expectedInstrumentationScopes,
  );

  return async () => {
    await toolkit.destroy(source, {
      stacks: {
        strategy: StackSelectionStrategy.ALL_STACKS,
      },
    });
  };
}
