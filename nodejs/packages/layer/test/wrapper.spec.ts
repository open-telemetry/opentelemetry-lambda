import type { AwsSdkInstrumentationConfig } from '@opentelemetry/instrumentation-aws-sdk';
import { SDKRegistrationConfig } from '@opentelemetry/sdk-trace-base';
import { TextMapPropagator } from '@opentelemetry/api';

import { wrap, unwrap } from '../src/wrapper';

import { stub } from 'sinon';
import assert from 'assert';

declare global {
  function configureAwsInstrumentation(
    defaultConfig: AwsSdkInstrumentationConfig,
  ): AwsSdkInstrumentationConfig;
  function configureSdkRegistration(
    defaultSdkRegistration: SDKRegistrationConfig,
  ): SDKRegistrationConfig;
}

describe('wrapper', () => {
  let oldEnv: NodeJS.ProcessEnv;

  beforeEach(() => {
    oldEnv = { ...process.env };

    unwrap();
  });

  afterEach(() => {
    process.env = oldEnv;

    unwrap();
  });

  describe('configureAwsInstrumentation', () => {
    it('is used if defined', () => {
      const configureAwsInstrumentationStub = stub().returns({
        suppressInternalInstrumentation: true,
      });
      global.configureAwsInstrumentation = configureAwsInstrumentationStub;
      wrap();
      assert(configureAwsInstrumentationStub.calledOnce);
    });
  });

  describe('getPropagator', () => {
    const testConfiguredPropagator = (
      propagatorNames: string[],
      expectedPropagatorFields: string[],
    ) => {
      if (propagatorNames && propagatorNames.length) {
        process.env.OTEL_PROPAGATORS = propagatorNames.join(',');
      }

      const configureSdkRegistrationStub = stub().returnsArg(0);
      global.configureSdkRegistration = configureSdkRegistrationStub;
      wrap();
      assert(configureSdkRegistrationStub.calledOnce);

      const sdkRegistrationConfig: SDKRegistrationConfig =
        configureSdkRegistrationStub.getCall(0).firstArg;
      assert.notEqual(sdkRegistrationConfig, null);

      const propagator: TextMapPropagator | null | undefined =
        sdkRegistrationConfig.propagator;
      assert.notEqual(propagator, null);

      const actualPropagatorFields: string[] | undefined = propagator?.fields();
      assert.notEqual(actualPropagatorFields, null);
      assert.deepEqual(actualPropagatorFields, expectedPropagatorFields);
    };

    it('is configured by default', () => {
      // by default, 'W3CTraceContextPropagator' and 'W3CBaggagePropagator' propagators are added.
      // - 'traceparent' and 'tracestate' fields are used by the 'W3CTraceContextPropagator'
      // - 'baggage' field is used by the 'W3CBaggagePropagator'
      testConfiguredPropagator([], ['traceparent', 'tracestate', 'baggage']);
    });

    it('is configured to w3c-trace-context by env var', () => {
      // 'traceparent' and 'tracestate' fields are used by the 'W3CTraceContextPropagator'
      testConfiguredPropagator(['tracecontext'], ['traceparent', 'tracestate']);
    });

    it('is configured to w3c-baggage by env var', () => {
      // 'baggage' field is used by the 'W3CBaggagePropagator'
      testConfiguredPropagator(['baggage'], ['baggage']);
    });

    it('is configured to xray by env var', () => {
      // 'x-amzn-trace-id' field is used by the 'AWSXRayPropagator'
      testConfiguredPropagator(['xray'], ['x-amzn-trace-id']);
    });

    it('is configured to xray-lambda by env var', () => {
      // 'x-amzn-trace-id' field is used by the 'AWSXRayLambdaPropagator'
      testConfiguredPropagator(['xray'], ['x-amzn-trace-id']);
    });

    it('is configured by unsupported propagator', () => {
      // in case of unsupported propagator, warning log is printed and empty propagator array is returned
      testConfiguredPropagator(['jaeger'], []);
    });
  });
});
