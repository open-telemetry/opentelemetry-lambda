// Example Stack
import { AssetType, TerraformAsset, TerraformOutput } from "cdktf";
import { Construct } from "constructs";
import path from "path";
import { LambdaLayerVersion } from "../.gen/providers/aws";
import { BaseStack, BaseStackConfig, Environment } from "./core";

export enum LambdaRuntime {
  Node14 = "nodejs14.x",
  Node12 = "nodejs12.x",
  Node10 = "nodejs10.x",
}

export interface LayerStackConfig extends BaseStackConfig {
  readonly environment: Environment;
  readonly nodeRuntimes: LambdaRuntime[]
}

export class LayerStack extends BaseStack {
  constructor(scope: Construct, id: string, config: LayerStackConfig) {
    super(scope, id, config);

    // Collector layer
    const collectorZipFile = new TerraformAsset(this, 'collector_zip', {
      path: path.resolve(__dirname, `../collector/build/collector-extension.zip`),
      type: AssetType.FILE,
    });

    const collectorLayer = new LambdaLayerVersion(this, 'collector_lambda_layer', {
      layerName: 'opentelemetry-collector',
      compatibleRuntimes: config.nodeRuntimes,
      filename: collectorZipFile.path,
    })

    new TerraformOutput(this, 'collector_layer_arn', {
      value: collectorLayer.arn,
    });

    // Node wrapper layer
    const nodeWrapperZipFile = new TerraformAsset(this, 'node_wrapper_zip', {
      path: path.resolve(__dirname, `../nodejs/packages/layer/build/layer.zip`),
      type: AssetType.FILE,
    });

    const nodeWrapperLayer = new LambdaLayerVersion(this, 'node_wrapper_lambda_layer', {
      layerName: 'opentelemetry-node-wrapper',
      compatibleRuntimes: config.nodeRuntimes,
      filename: nodeWrapperZipFile.path,
    })

    new TerraformOutput(this, 'node_wrapper_layer_arn', {
      value: nodeWrapperLayer.arn,
    });
  }
}