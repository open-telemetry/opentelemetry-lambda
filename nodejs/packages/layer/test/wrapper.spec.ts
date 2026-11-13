import { init, wrap, unwrap } from '../src/wrapper';

import {
  propagation,
  ProxyTracerProvider,
  ROOT_CONTEXT,
  trace,
  TraceFlags,
} from '@opentelemetry/api';
import type { AwsSdkInstrumentationConfig } from '@opentelemetry/instrumentation-aws-sdk';
import { TRACE_PARENT_HEADER } from '@opentelemetry/core';
import {
  AWSXRAY_TRACE_ID_HEADER,
  AWSXRayPropagator,
} from '@opentelemetry/propagator-aws-xray';
import {
  ConsoleSpanExporter,
  InMemorySpanExporter,
  SimpleSpanProcessor,
} from '@opentelemetry/sdk-trace';
import { resourceFromAttributes } from '@opentelemetry/resources';
import {
  AggregationTemporality,
  InMemoryMetricExporter,
  PeriodicExportingMetricReader,
} from '@opentelemetry/sdk-metrics';
import { OTLPTraceExporter } from '@opentelemetry/exporter-trace-otlp-http';

import { stub } from 'sinon';
import assert from 'assert';

declare global {
  function configureAwsInstrumentation(
    defaultConfig: AwsSdkInstrumentationConfig,
  ): AwsSdkInstrumentationConfig;
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
      await setConfiguredPropagator(propagatorNames);
      assert.deepEqual(propagation.fields(), expectedPropagatorFields);
    };

    const setConfiguredPropagator = async (propagatorNames: string[]) => {
      if (propagatorNames && propagatorNames.length) {
        process.env.OTEL_PROPAGATORS = propagatorNames.join(',');
      }
      await wrap();
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

    it('is configured to b3 by env var', async () => {
      // 'b3' field is used by the single header 'B3Propagator'
      await testConfiguredPropagator(['b3'], ['b3']);
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

      await setConfiguredPropagator(['tracecontext', 'xray']);
      const extractedSpanContext1 = trace
        .getSpan(propagation.extract(ROOT_CONTEXT, carrier))
        ?.spanContext();
      // Last one overwrites, so we will see the context extracted from the last propagator (xray)
      assert.deepStrictEqual(extractedSpanContext1, {
        traceId: AWS_XRAY_TRACE_ID.replace('-', ''),
        spanId: AWS_XRAY_SPAN_ID,
        isRemote: true,
        traceFlags: TraceFlags.SAMPLED,
      });

      await unwrap();

      await setConfiguredPropagator(['xray', 'tracecontext']);
      const extractedSpanContext2 = trace
        .getSpan(propagation.extract(ROOT_CONTEXT, carrier))
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

  describe('configure hooks', () => {
    after(() => {
      delete (global as any).configureTracer;
      delete (global as any).configureMeter;
      delete (global as any).configureLogger;
      delete (global as any).configureSdkRegistration;
    });

    it('use the propagator from configureSdkRegistration', async () => {
      process.env.OTEL_PROPAGATORS = 'tracecontext';
      global.configureSdkRegistration = config => ({
        ...config,
        propagator: new AWSXRayPropagator(),
      });
      await wrap();

      assert.deepEqual(propagation.fields(), ['x-amzn-trace-id']);
    });

    it('use OTEL_PROPAGATORS when configureSdkRegistration sets no propagator', async () => {
      process.env.OTEL_PROPAGATORS = 'xray';
      global.configureSdkRegistration = config => config;
      await wrap();

      assert.deepEqual(propagation.fields(), ['x-amzn-trace-id']);
    });

    it('merge the resource from configureTracer with the detected one', async () => {
      const exporter = new InMemorySpanExporter();
      global.configureTracer = config => ({
        ...config,
        resource: resourceFromAttributes({ 'test.attribute': 'from-hook' }),
        spanProcessors: [new SimpleSpanProcessor({ exporter })],
      });
      delete (global as any).configureMeter;
      delete (global as any).configureLogger;
      await wrap();

      trace.getTracer('test').startSpan('span').end();
      const [span] = exporter.getFinishedSpans();
      assert.strictEqual(
        span.resource.attributes['test.attribute'],
        'from-hook',
      );
      // Set by the process detector.
      assert.strictEqual(span.resource.attributes['process.pid'], process.pid);
    });

    it('pass readers to NodeSDK when configureMeter adds them', async () => {
      const reader = new PeriodicExportingMetricReader({
        exporter: new InMemoryMetricExporter(AggregationTemporality.CUMULATIVE),
      });
      global.configureTracer = config => config;
      global.configureMeter = (config: any) => ({
        ...config,
        readers: [reader],
      });
      delete (global as any).configureLogger;
      await wrap();

      // collect() throws if NodeSDK did not register the reader.
      await reader.collect();
      await reader.shutdown();
    });
  });

  describe('exporters', () => {
    before(() => {
      // TODO: Does this belong here
      delete (global as any).configureTracer;
    });

    beforeEach(() => {
      // NodeSDK defaults to http/protobuf. Ask for the http/json exporter the layer depends on.
      process.env.OTEL_EXPORTER_OTLP_TRACES_PROTOCOL = 'http/json';
    });

    const testConfiguredExporter = async (
      exporterNames: string[] | undefined,
      expectedExporters: any[],
    ) => {
      if (exporterNames) {
        process.env.OTEL_TRACES_EXPORTER = exporterNames.join(',');
      }

      await wrap();
      const tracerProvider = (
        trace.getTracerProvider() as ProxyTracerProvider
      ).getDelegate() as any;
      const spanProcessors =
        tracerProvider._activeSpanProcessor._spanProcessors;

      assert.strictEqual(spanProcessors.length, expectedExporters.length);
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
