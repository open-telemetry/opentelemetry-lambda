import { init, wrap, unwrap } from '../src/wrapper';

import {
  context,
  createContextKey,
  logs,
  metrics,
  propagation,
  ProxyTracerProvider,
  ROOT_CONTEXT,
  trace,
  TraceFlags,
} from '@opentelemetry/api';
import type { AwsSdkInstrumentationConfig } from '@opentelemetry/instrumentation-aws-sdk';
import {
  TRACE_PARENT_HEADER,
  W3CTraceContextPropagator,
} from '@opentelemetry/core';
import {
  AWSXRAY_TRACE_ID_HEADER,
  AWSXRayPropagator,
} from '@opentelemetry/propagator-aws-xray';
import {
  AlwaysOffSampler,
  ConsoleSpanExporter,
  InMemorySpanExporter,
  SimpleSpanProcessor,
} from '@opentelemetry/sdk-trace';
import { resourceFromAttributes } from '@opentelemetry/resources';
import { AsyncLocalStorageContextManager } from '@opentelemetry/context-async-hooks';
import {
  AggregationTemporality,
  ConsoleMetricExporter,
  InMemoryMetricExporter,
  MeterProvider,
  PeriodicExportingMetricReader,
} from '@opentelemetry/sdk-metrics';
import { OTLPTraceExporter } from '@opentelemetry/exporter-trace-otlp-http';
import { OTLPMetricExporter as HttpOTLPMetricExporter } from '@opentelemetry/exporter-metrics-otlp-http';
import { OTLPMetricExporter as ProtoOTLPMetricExporter } from '@opentelemetry/exporter-metrics-otlp-proto';
import {
  InMemoryLogRecordExporter,
  SimpleLogRecordProcessor,
} from '@opentelemetry/sdk-logs';

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
    afterEach(() => {
      delete (global as any).configureTracer;
      delete (global as any).configureMeter;
      delete (global as any).configureLogger;
      delete (global as any).configureSdkRegistration;
      delete (global as any).configureSdk;
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

    it('leave the context manager unset when configureSdkRegistration returns null', async () => {
      global.configureSdkRegistration = config => ({
        ...config,
        contextManager: null,
      });
      await wrap();

      const key = createContextKey('test');
      const ctx = ROOT_CONTEXT.setValue(key, 'value');
      assert.strictEqual(
        context.with(ctx, () => context.active()),
        ROOT_CONTEXT,
      );
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

    describe('metric readers', () => {
      beforeEach(() => {
        process.env.OTEL_METRICS_EXPORTER = 'otlp';
        process.env.OTEL_TRACES_EXPORTER = 'none';
        process.env.OTEL_LOGS_EXPORTER = 'none';
      });

      afterEach(async () => {
        const provider = metrics.getMeterProvider();
        if (provider instanceof MeterProvider) {
          await provider.shutdown();
        }
      });

      it('disable env readers when configureMeter returns a new empty list', async () => {
        global.configureMeter = () => ({ readers: [] });
        await wrap();

        assert.ok(!(metrics.getMeterProvider() instanceof MeterProvider));
      });

      it('use env readers when configureMeter returns the unchanged config', async () => {
        global.configureMeter = config => config;
        await wrap();

        assert.ok(metrics.getMeterProvider() instanceof MeterProvider);
      });

      it('use env readers when configureMeter copies the unchanged list', async () => {
        global.configureMeter = (config: any) => ({ ...config });
        await wrap();

        assert.ok(metrics.getMeterProvider() instanceof MeterProvider);
      });

      it('keep an empty reader list from configureSdk', async () => {
        global.configureSdk = config => ({ ...config, metricReaders: [] });
        global.configureMeter = config => config;
        await wrap();

        assert.ok(!(metrics.getMeterProvider() instanceof MeterProvider));
      });

      it('keep the OTLP reader when configureMeter adds a reader', async () => {
        const reader = new PeriodicExportingMetricReader({
          exporter: new InMemoryMetricExporter(
            AggregationTemporality.CUMULATIVE,
          ),
        });
        let defaults: any[] = [];
        global.configureMeter = (config: any) => {
          defaults = [...config.readers];
          config.readers.push(reader);
          return config;
        };
        await wrap();

        assert.strictEqual(defaults.length, 1);
        assert.ok(
          defaults[0]._exporter instanceof ProtoOTLPMetricExporter,
          'default protocol is http/protobuf',
        );
        // collect() throws if NodeSDK did not register the reader.
        await defaults[0].collect();
        await reader.collect();
      });

      it('give configureMeter the env readers for the OTLP protocol', async () => {
        process.env.OTEL_METRICS_EXPORTER = 'otlp,console';
        process.env.OTEL_EXPORTER_OTLP_METRICS_PROTOCOL = 'http/json';
        let defaults: any[] = [];
        global.configureMeter = (config: any) => {
          defaults = config.readers;
          return config;
        };
        await wrap();

        assert.strictEqual(defaults.length, 2);
        assert.ok(defaults[0]._exporter instanceof HttpOTLPMetricExporter);
        assert.ok(defaults[1]._exporter instanceof ConsoleMetricExporter);
      });

      it('give configureMeter no readers when OTEL_METRICS_EXPORTER is none', async () => {
        process.env.OTEL_METRICS_EXPORTER = 'none';
        let defaults: any[] | undefined;
        global.configureMeter = (config: any) => {
          defaults = config.readers;
          return config;
        };
        await wrap();

        assert.deepStrictEqual(defaults, []);
        assert.ok(!(metrics.getMeterProvider() instanceof MeterProvider));
      });

      it('register readers added to the placeholder list', async () => {
        const reader = new PeriodicExportingMetricReader({
          exporter: new InMemoryMetricExporter(
            AggregationTemporality.CUMULATIVE,
          ),
        });
        global.configureMeter = (config: any) => {
          config.readers.push(reader);
          return config;
        };
        await wrap();
        await reader.collect();
      });
    });

    it('create NodeSDK with the options from configureSdk', async () => {
      const exporter = new InMemorySpanExporter();
      delete (global as any).configureTracer;
      delete (global as any).configureMeter;
      delete (global as any).configureLogger;
      delete (global as any).configureSdkRegistration;
      global.configureSdk = config => ({
        ...config,
        idGenerator: {
          generateTraceId: () => 'f'.repeat(32),
          generateSpanId: () => 'e'.repeat(16),
        },
        spanProcessors: [new SimpleSpanProcessor({ exporter })],
      });
      await wrap();

      trace.getTracer('test').startSpan('span').end();
      const [span] = exporter.getFinishedSpans();
      assert.strictEqual(span.spanContext().traceId, 'f'.repeat(32));
      // The layer's resource detectors are still in the config.
      assert.strictEqual(span.resource.attributes['process.pid'], process.pid);
    });

    it('keep an empty processor list from configureSdk through configureTracer', async () => {
      process.env.OTEL_TRACES_EXPORTER = 'console';
      global.configureSdk = config => ({ ...config, spanProcessors: [] });
      global.configureTracer = config => config;
      await wrap();

      const span = trace.getTracer('test').startSpan('disabled');
      assert.strictEqual(span.isRecording(), false);
      span.end();
    });

    it('apply configureSdkRegistration after configureSdk', async () => {
      process.env.OTEL_PROPAGATORS = 'tracecontext';
      global.configureSdk = config => ({
        ...config,
        textMapPropagator: new AWSXRayPropagator(),
      });
      global.configureSdkRegistration = config => {
        assert.deepEqual(config.propagator?.fields(), ['x-amzn-trace-id']);
        return { ...config, propagator: new W3CTraceContextPropagator() };
      };
      await wrap();

      assert.deepEqual(propagation.fields(), ['traceparent', 'tracestate']);
    });

    it('let configureTracer extend the options from configureSdk', async () => {
      const exporter = new InMemorySpanExporter();
      global.configureSdk = config => ({
        ...config,
        resource: resourceFromAttributes({ 'test.sdk': 'kept' }),
        idGenerator: {
          generateTraceId: () => 'f'.repeat(32),
          generateSpanId: () => 'e'.repeat(16),
        },
        spanProcessors: [new SimpleSpanProcessor({ exporter })],
      });
      global.configureTracer = config => ({
        ...config,
        resource: config.resource!.merge(
          resourceFromAttributes({ 'test.tracer': 'added' }),
        ),
      });
      await wrap();

      trace.getTracer('test').startSpan('span').end();
      const [span] = exporter.getFinishedSpans();
      assert.strictEqual(span.resource.attributes['test.sdk'], 'kept');
      assert.strictEqual(span.resource.attributes['test.tracer'], 'added');
      assert.strictEqual(span.spanContext().traceId, 'f'.repeat(32));
    });

    it('let configureMeter add to readers from configureSdk', async () => {
      const readers = [0, 1].map(
        () =>
          new PeriodicExportingMetricReader({
            exporter: new InMemoryMetricExporter(
              AggregationTemporality.CUMULATIVE,
            ),
          }),
      );
      global.configureSdk = config => ({
        ...config,
        metricReaders: [readers[0]],
      });
      global.configureMeter = (config: any) => ({
        ...config,
        readers: [...config.readers, readers[1]],
      });
      try {
        await wrap();
        for (const reader of readers) {
          // collect() throws if the reader was not registered with NodeSDK.
          await reader.collect();
        }
      } finally {
        await Promise.all(readers.map(reader => reader.shutdown()));
      }
    });

    it('let configureLogger add to processors from configureSdk', async () => {
      const exporters = [
        new InMemoryLogRecordExporter(),
        new InMemoryLogRecordExporter(),
      ];
      const processors = exporters.map(
        exporter => new SimpleLogRecordProcessor({ exporter }),
      );
      global.configureSdk = config => ({
        ...config,
        logRecordProcessors: [processors[0]],
      });
      global.configureLogger = (config: any) => ({
        ...config,
        processors: [...config.processors, processors[1]],
      });
      try {
        await wrap();
        logs.getLogger('test').emit({ body: 'from-both-hooks' });
        await Promise.all(processors.map(processor => processor.forceFlush()));
        for (const exporter of exporters) {
          assert.deepEqual(
            exporter.getFinishedLogRecords().map(record => record.body),
            ['from-both-hooks'],
          );
        }
      } finally {
        await Promise.all(processors.map(processor => processor.shutdown()));
      }
    });

    it('keep the sampler configureTracer leaves out', async () => {
      const exporter = new InMemorySpanExporter();
      global.configureSdk = config => ({
        ...config,
        sampler: new AlwaysOffSampler(),
      });
      global.configureTracer = () => ({
        spanProcessors: [new SimpleSpanProcessor({ exporter })],
      });
      await wrap();

      const span = trace.getTracer('test').startSpan('unsampled');
      assert.strictEqual(span.isRecording(), false);
      span.end();
      assert.deepEqual(exporter.getFinishedSpans(), []);
    });

    it('keep the resource, ID generator and span limits configureTracer leaves out', async () => {
      const exporter = new InMemorySpanExporter();
      global.configureSdk = config => ({
        ...config,
        resource: resourceFromAttributes({ 'test.sdk': 'kept' }),
        idGenerator: {
          generateTraceId: () => 'f'.repeat(32),
          generateSpanId: () => 'e'.repeat(16),
        },
        spanLimits: { attributeCountLimit: 1 },
      });
      global.configureTracer = () => ({
        spanProcessors: [new SimpleSpanProcessor({ exporter })],
      });
      await wrap();

      const span = trace.getTracer('test').startSpan('span');
      span.setAttribute('first', 'kept');
      span.setAttribute('second', 'dropped');
      span.end();
      const [finishedSpan] = exporter.getFinishedSpans();
      assert.strictEqual(finishedSpan.resource.attributes['test.sdk'], 'kept');
      assert.strictEqual(finishedSpan.spanContext().traceId, 'f'.repeat(32));
      assert.strictEqual(finishedSpan.spanContext().spanId, 'e'.repeat(16));
      assert.deepEqual(finishedSpan.attributes, { first: 'kept' });
    });

    it('keep SDK readers when configureMeter returns null', async () => {
      const reader = new PeriodicExportingMetricReader({
        exporter: new InMemoryMetricExporter(AggregationTemporality.CUMULATIVE),
      });
      global.configureSdk = config => ({ ...config, metricReaders: [reader] });
      global.configureMeter = () => null;
      try {
        await wrap();
        // collect() throws if NodeSDK did not register the reader.
        await reader.collect();
      } finally {
        await reader.shutdown();
      }
    });

    it('keep SDK processors when configureLogger returns undefined', async () => {
      const exporter = new InMemoryLogRecordExporter();
      const processor = new SimpleLogRecordProcessor({ exporter });
      global.configureSdk = config => ({
        ...config,
        logRecordProcessors: [processor],
      });
      global.configureLogger = () => undefined;
      try {
        await wrap();
        logs.getLogger('test').emit({ body: 'from-sdk' });
        await processor.forceFlush();
        assert.deepEqual(
          exporter.getFinishedLogRecords().map(record => record.body),
          ['from-sdk'],
        );
      } finally {
        await processor.shutdown();
      }
    });

    it('keep the SDK context manager when configureSdkRegistration returns only a propagator', async () => {
      const contextManager = new AsyncLocalStorageContextManager().enable();
      global.configureSdk = config => ({ ...config, contextManager });
      global.configureSdkRegistration = () => ({
        propagator: new AWSXRayPropagator(),
      });
      try {
        await wrap();
        const key = createContextKey('test');
        const ctx = ROOT_CONTEXT.setValue(key, 'value');
        assert.strictEqual(
          contextManager.with(ctx, () => context.active()),
          ctx,
        );
        assert.deepEqual(propagation.fields(), ['x-amzn-trace-id']);
      } finally {
        contextManager.disable();
      }
    });

    it('use NodeSDK options the other hooks do not offer', async () => {
      const exporter = new InMemorySpanExporter();
      delete (global as any).configureSdkRegistration;
      global.configureSdk = config => ({
        ...config,
        serviceName: 'from-configure-sdk',
        spanProcessors: [new SimpleSpanProcessor({ exporter })],
      });
      await wrap();

      trace.getTracer('test').startSpan('span').end();
      const [span] = exporter.getFinishedSpans();
      assert.strictEqual(
        span.resource.attributes['service.name'],
        'from-configure-sdk',
      );
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

    it('uses env exporters when configureTracer leaves its default list empty', async () => {
      global.configureTracer = config => config;
      try {
        await testConfiguredExporter(['console'], [ConsoleSpanExporter]);
      } finally {
        delete (global as any).configureTracer;
      }
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
