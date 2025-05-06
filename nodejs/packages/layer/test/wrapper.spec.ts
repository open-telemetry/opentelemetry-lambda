import { init, wrap, unwrap } from '../src/wrapper';

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
import {
  ConsoleSpanExporter,
  NodeTracerProvider,
  SDKRegistrationConfig,
} from '@opentelemetry/sdk-trace-node';
import { OTLPTraceExporter } from '@opentelemetry/exporter-trace-otlp-http';

import { SinonSpy, spy, stub } from 'sinon';
import assert from 'assert';

declare global {
  function configureAwsInstrumentation(
    defaultConfig: AwsSdkInstrumentationConfig,
  ): AwsSdkInstrumentationConfig;
  function configureSdkRegistration(
    defaultSdkRegistration: SDKRegistrationConfig,
  ): SDKRegistrationConfig;
}

describe('wrapper', async () => {
  let oldEnv: NodeJS.ProcessEnv;

  await init();

  beforeEach(async () => {
    oldEnv = { ...process.env };

    await unwrap();
  });

  afterEach(async () => {
    process.env = oldEnv;

    await unwrap();
  });

  describe('configureAwsInstrumentation', () => {
    it('is used if defined', async () => {
      const configureAwsInstrumentationStub = stub().returns({
        suppressInternalInstrumentation: true,
      });
      global.configureAwsInstrumentation = configureAwsInstrumentationStub;
      await wrap();
      assert(configureAwsInstrumentationStub.calledOnce);
    });
  });

  describe('getPropagator', () => {
    const testConfiguredPropagator = async (
      propagatorNames: string[],
      expectedPropagatorFields: string[],
    ) => {
      if (propagatorNames && propagatorNames.length) {
        process.env.OTEL_PROPAGATORS = propagatorNames.join(',');
      }

      const configureSdkRegistrationStub = stub().returnsArg(0);
      global.configureSdkRegistration = configureSdkRegistrationStub;
      await wrap();
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

    const setAndGetConfiguredPropagator = async (
      propagatorNames: string[],
    ): Promise<TextMapPropagator> => {
      if (propagatorNames && propagatorNames.length) {
        process.env.OTEL_PROPAGATORS = propagatorNames.join(',');
      }

      const configureSdkRegistrationStub = stub().returnsArg(0);
      global.configureSdkRegistration = configureSdkRegistrationStub;
      await wrap();
      assert(configureSdkRegistrationStub.calledOnce);

      const sdkRegistrationConfig: SDKRegistrationConfig =
        configureSdkRegistrationStub.getCall(0).firstArg;
      assert.notEqual(sdkRegistrationConfig, null);

      const propagator: TextMapPropagator | null | undefined =
        sdkRegistrationConfig.propagator;
      assert.notEqual(propagator, null);

      return propagator!;
    };

    it('is configured by default', async () => {
      // by default, 'W3CTraceContextPropagator' and 'W3CBaggagePropagator' propagators are added.
      // - 'traceparent' and 'tracestate' fields are used by the 'W3CTraceContextPropagator'
      // - 'baggage' field is used by the 'W3CBaggagePropagator'
      await testConfiguredPropagator(
        [],
        ['traceparent', 'tracestate', 'baggage'],
      );
    });

    it('is configured to w3c-trace-context by env var', async () => {
      // 'traceparent' and 'tracestate' fields are used by the 'W3CTraceContextPropagator'
      await testConfiguredPropagator(
        ['tracecontext'],
        ['traceparent', 'tracestate'],
      );
    });

    it('is configured to w3c-baggage by env var', async () => {
      // 'baggage' field is used by the 'W3CBaggagePropagator'
      await testConfiguredPropagator(['baggage'], ['baggage']);
    });

    it('is configured to xray by env var', async () => {
      // 'x-amzn-trace-id' field is used by the 'AWSXRayPropagator'
      await testConfiguredPropagator(['xray'], ['x-amzn-trace-id']);
    });

    it('is configured to xray-lambda by env var', async () => {
      // 'x-amzn-trace-id' field is used by the 'AWSXRayLambdaPropagator'
      await testConfiguredPropagator(['xray'], ['x-amzn-trace-id']);
    });

    it('is configured by unsupported propagator', async () => {
      // in case of unsupported propagator, warning log is printed and empty propagator array is returned
      await testConfiguredPropagator(['jaeger'], []);
    });

    it('is configured in correct order', async () => {
      const W3C_TRACE_ID = '5b8aa5a2d2c872e8321cf37308d69df2';
      const W3C_SPAN_ID = '051581bf3cb55c13';
      const AWS_XRAY_TRACE_ID = '8a3c60f7-d188f8fa79d48a391a778fa6';
      const AWS_XRAY_SPAN_ID = '53995c3f42cd8ad8';
      const carrier = {
        [TRACE_PARENT_HEADER]: `00-${W3C_TRACE_ID}-${W3C_SPAN_ID}-01`,
        [AWSXRAY_TRACE_ID_HEADER]: `Root=1-${AWS_XRAY_TRACE_ID};Parent=${AWS_XRAY_SPAN_ID};Sampled=1`,
      };

      const propagator1: TextMapPropagator =
        await setAndGetConfiguredPropagator(['tracecontext', 'xray']);
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

      await unwrap();

      const propagator2: TextMapPropagator =
        await setAndGetConfiguredPropagator(['xray', 'tracecontext']);
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

  describe('exporters', () => {
    let providerSpy: SinonSpy;

    before(() => {
      // TODO: Does this belong here
      delete (global as any).configureTracer;
    });

    beforeEach(() => {
      providerSpy = spy(NodeTracerProvider.prototype, 'register');
    });

    afterEach(() => {
      providerSpy.restore();
    });

    const testConfiguredExporter = async (
      exporterNames: string[] | undefined,
      expectedExporters: any[],
    ) => {
      if (exporterNames) {
        process.env.OTEL_TRACES_EXPORTER = exporterNames.join(',');
      }

      await wrap();
      assert(providerSpy.calledOnce);
      const tracer = providerSpy.getCall(0).thisValue;
      const spanProcessors = tracer._config.spanProcessors;

      for (const [i, processor] of spanProcessors.entries()) {
        assert.ok(processor._exporter instanceof expectedExporters[i]);
      }
    };

    it('is configured to OTLP by default', async () => {
      await testConfiguredExporter(undefined, [OTLPTraceExporter]);
    });

    it('is configured to console by env var', async () => {
      await testConfiguredExporter(['console'], [ConsoleSpanExporter]);
    });

    it('is configured to both console and otlp by env var', async () => {
      await testConfiguredExporter(
        ['console', 'otlp'],
        [ConsoleSpanExporter, OTLPTraceExporter],
      );
    });
  });
});
