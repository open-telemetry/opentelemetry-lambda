import { wrap, unwrap } from '../src/wrapper';

import {
  defaultTextMapGetter,
  ROOT_CONTEXT,
  TextMapPropagator,
  trace,
  TraceFlags,
} from '@opentelemetry/api';
import type { AwsSdkInstrumentationConfig } from '@opentelemetry/instrumentation-aws-sdk';
import { TRACE_PARENT_HEADER } from '@opentelemetry/core';
import { AWSXRAY_TRACE_ID_HEADER } from '@opentelemetry/propagator-aws-xray';
import { SDKRegistrationConfig } from '@opentelemetry/sdk-trace-base';

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

    const setAndGetConfiguredPropagator = (
      propagatorNames: string[],
    ): TextMapPropagator => {
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

      return propagator!;
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

    it('is configured in correct order', () => {
      const W3C_TRACE_ID = '5b8aa5a2d2c872e8321cf37308d69df2';
      const W3C_SPAN_ID = '051581bf3cb55c13';
      const AWS_XRAY_TRACE_ID = '8a3c60f7-d188f8fa79d48a391a778fa6';
      const AWS_XRAY_SPAN_ID = '53995c3f42cd8ad8';
      const carrier = {
        [TRACE_PARENT_HEADER]: `00-${W3C_TRACE_ID}-${W3C_SPAN_ID}-01`,
        [AWSXRAY_TRACE_ID_HEADER]: `Root=1-${AWS_XRAY_TRACE_ID};Parent=${AWS_XRAY_SPAN_ID};Sampled=1`,
      };

      const propagator1: TextMapPropagator = setAndGetConfiguredPropagator([
        'tracecontext',
        'xray',
      ]);
      const extractedSpanContext1 = trace
        .getSpan(
          propagator1.extract(ROOT_CONTEXT, carrier, defaultTextMapGetter),
        )
        ?.spanContext();
      // Last one overwrites, so we will see the context extracted from the last propagator (xray)
      assert.deepStrictEqual(extractedSpanContext1, {
        traceId: AWS_XRAY_TRACE_ID.replace('-', ''),
        spanId: AWS_XRAY_SPAN_ID,
        isRemote: true,
        traceFlags: TraceFlags.SAMPLED,
      });

      unwrap();

      const propagator2: TextMapPropagator = setAndGetConfiguredPropagator([
        'xray',
        'tracecontext',
      ]);
      const extractedSpanContext2 = trace
        .getSpan(
          propagator2.extract(ROOT_CONTEXT, carrier, defaultTextMapGetter),
        )
        ?.spanContext();
      // Last one overwrites, so we will see the context extracted from the last propagator (tracecontext)
      assert.deepStrictEqual(extractedSpanContext2, {
        traceId: W3C_TRACE_ID,
        spanId: W3C_SPAN_ID,
        isRemote: true,
        traceFlags: TraceFlags.SAMPLED,
      });
    });
  });
});
