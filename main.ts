import { CommonConfig, Environment, Region, Namespace } from "./lib/core";
import { LambdaRuntime, LayerStack, LayerStackConfig } from "./lib/layerStack";
import { App } from "cdktf";

const tags = new Map<string, string>([
  ["owner", "platform-infra"],
]);

const common: CommonConfig = {
  serviceName: 'opentelemetry-lambda',
  region: Region.US_EAST_1,
  namespace: Namespace.Medly,
  tags: tags,
}

const envs: LayerStackConfig[] = [
  {
    ...common,
    environment: Environment.Development,
    nodeRuntimes: [LambdaRuntime.Node14, LambdaRuntime.Node12, LambdaRuntime.Node10],
  },
  {
    ...common,
    environment: Environment.Production,
    nodeRuntimes: [LambdaRuntime.Node14, LambdaRuntime.Node12, LambdaRuntime.Node10],
  }
]

const app = new App();
envs.map(env => new LayerStack(app, env.environment, env));
app.synth();