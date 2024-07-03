import type { AwsSdkInstrumentationConfig } from "@opentelemetry/instrumentation-aws-sdk";
import { stub } from "sinon";

declare global {
  function configureAwsInstrumentation(
    defaultConfig: AwsSdkInstrumentationConfig,
  ): AwsSdkInstrumentationConfig;
}

const assert = require("assert");

describe("wrapper", () => {
  describe("configureAwsInstrumentation", () => {
    it("is used if defined", () => {
      const configureAwsInstrumentationStub = stub().returns({
        suppressInternalInstrumentation: true,
      });
      global.configureAwsInstrumentation = configureAwsInstrumentationStub;
      require("../src/wrapper");
      assert(configureAwsInstrumentationStub.calledOnce);
    }).timeout(10000);
  });
});
