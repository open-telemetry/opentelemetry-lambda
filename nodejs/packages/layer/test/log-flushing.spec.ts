import assert from 'assert';
import { Context } from 'aws-lambda';
import { diag, metrics } from '@opentelemetry/api';
import { logs } from '@opentelemetry/api-logs';
import {
  BatchLogRecordProcessor,
  InMemoryLogRecordExporter,
  LoggerProvider,
} from '@opentelemetry/sdk-logs';
import { MeterProvider } from '@opentelemetry/sdk-metrics';
import {
  BatchSpanProcessor,
  InMemorySpanExporter,
  NodeTracerProvider,
} from '@opentelemetry/sdk-trace-node';
import { restore, spy, stub } from 'sinon';
import { init, wrap, unwrap } from '../src/wrapper';

describe('invocation log flushing', () => {
  let oldEnv: NodeJS.ProcessEnv;
  let oldHooks: Record<string, unknown>;
  let exporter: InMemoryLogRecordExporter;
  let spanExporter: InMemorySpanExporter;
  let loggerProvider: LoggerProvider | undefined;
  let meterProvider: MeterProvider;
  const hooks = [
    'configureTracer',
    'configureSdkRegistration',
    'configureMeter',
    'configureMeterProvider',
    'configureLogger',
    'configureLoggerProvider',
    'configureLambdaInstrumentation',
    'configureInstrumentations',
  ];
  const ctx = {
    functionName: 'my_function',
    invokedFunctionArn: 'my_arn',
    awsRequestId: 'aws_request_id',
  } as Context;

  before(async () => {
    await init();
  });

  beforeEach(async () => {
    await unwrap();
    oldEnv = { ...process.env };
    oldHooks = {};
    for (const hook of hooks) {
      oldHooks[hook] = (global as any)[hook];
      delete (global as any)[hook];
    }
    process.env.LAMBDA_TASK_ROOT = __dirname;
    process.env._HANDLER = './handler/cjs/logs.handler';
    process.env.AWS_EXECUTION_ENV = 'AWS_Lambda_nodejs22.x';
    delete process.env.OTEL_LOGS_EXPORTER;
    delete process.env.OTEL_TRACES_EXPORTER;
    delete process.env.OTEL_METRICS_EXPORTER;
    delete require.cache[require.resolve('./handler/cjs/logs.js')];

    loggerProvider = undefined;
    exporter = new InMemoryLogRecordExporter();
    // Finish exporting on a later turn to verify completion waits for delivery.
    const exportLogs = exporter.export.bind(exporter);
    stub(exporter, 'export').callsFake((records, callback) => {
      setImmediate(() => exportLogs(records, callback));
    });
    spanExporter = new InMemorySpanExporter();
    global.configureTracer = () => ({
      spanProcessors: [new BatchSpanProcessor(spanExporter)],
    });
    global.configureMeter = () => ({ readers: [] });
    global.configureMeterProvider = provider => {
      meterProvider = provider as MeterProvider;
      metrics.setGlobalMeterProvider(meterProvider);
    };
    global.configureLogger = config => ({
      ...(config as object),
      processors: [
        new BatchLogRecordProcessor({
          exporter,
          scheduledDelayMillis: 60000,
        }),
      ],
    });
    global.configureInstrumentations = () => [];
  });

  afterEach(async () => {
    restore();
    await loggerProvider?.shutdown();
    await meterProvider?.shutdown();
    await unwrap();
    process.env = oldEnv;
    for (const hook of hooks) {
      if (oldHooks[hook] === undefined) {
        delete (global as any)[hook];
      } else {
        (global as any)[hook] = oldHooks[hook];
      }
    }
  });

  const loadHandler = async () => {
    await wrap();
    if (process.env.OTEL_LOGS_EXPORTER !== 'none') {
      loggerProvider = logs.getLoggerProvider() as LoggerProvider;
    }
    return require('./handler/cjs/logs.js');
  };

  const assertLogs = (expected = ['handler log']) => {
    assert.deepStrictEqual(
      exporter.getFinishedLogRecords().map(record => record.body),
      expected,
    );
  };

  for (const runtime of [22, 24]) {
    for (const fails of [false, true]) {
      it(`flushes before a promise ${fails ? 'rejects' : 'resolves'} on Node.js ${runtime}`, async () => {
        process.env.AWS_EXECUTION_ENV = `AWS_Lambda_nodejs${runtime}.x`;
        const { handler } = await loadHandler();
        const error = fails ? new Error('handler failed') : undefined;
        const result = handler({ error }, ctx);
        assert.strictEqual(exporter.getFinishedLogRecords().length, 0);
        if (error) {
          await assert.rejects(result, actual => actual === error);
        } else {
          assert.strictEqual(await result, 'ok');
        }
        assertLogs();
        assert.strictEqual(spanExporter.getFinishedSpans().length, 1);
      });
    }
  }

  for (const fails of [false, true]) {
    it(`flushes before the ${fails ? 'error' : 'success'} callback`, async () => {
      const { handler } = await loadHandler();
      const error = fails ? new Error('handler failed') : undefined;
      await new Promise<void>((resolve, reject) => {
        handler(
          { callback: true, error },
          ctx,
          (err: unknown, result: string) => {
            try {
              assert.strictEqual(err, error);
              assert.strictEqual(result, 'ok');
              assertLogs(['handler log', 'callback log']);
              resolve();
            } catch (e) {
              reject(e);
            }
          },
        );
      });
    });

    it(`flushes before a streaming handler ${fails ? 'rejects' : 'resolves'}`, async () => {
      process.env._HANDLER = './handler/cjs/logs.streamingHandler';
      const { streamingHandler } = await loadHandler();
      const error = fails ? new Error('handler failed') : undefined;
      const result = streamingHandler({ error }, {}, ctx);
      if (error) {
        await assert.rejects(result, actual => actual === error);
      } else {
        await result;
      }
      assertLogs();
    });
  }

  it('flushes logs with tracing disabled', async () => {
    process.env.OTEL_TRACES_EXPORTER = 'none';
    const { handler } = await loadHandler();
    assert.strictEqual(await handler({}, ctx), 'ok');
    assertLogs();
    assert.strictEqual(spanExporter.getFinishedSpans().length, 0);
  });

  it('keeps trace and metric flushing when logs are disabled', async () => {
    process.env.OTEL_LOGS_EXPORTER = 'none';
    const traceFlush = spy(NodeTracerProvider.prototype, 'forceFlush');
    const metricFlush = spy(MeterProvider.prototype, 'forceFlush');
    const { handler } = await loadHandler();
    assert.strictEqual(await handler({}, ctx), 'ok');
    assert(traceFlush.calledOnce);
    assert(metricFlush.calledOnce);
    assert.strictEqual(exporter.getFinishedLogRecords().length, 0);
    assert.strictEqual(spanExporter.getFinishedSpans().length, 1);
  });

  it('flushes a provider configured through the legacy hook', async () => {
    const configureProvider = spy((provider: unknown) => {
      logs.setGlobalLoggerProvider(provider as LoggerProvider);
    });
    global.configureLoggerProvider = configureProvider;
    const { handler } = await loadHandler();
    assert.strictEqual(await handler({}, ctx), 'ok');
    assert(configureProvider.calledOnce);
    assertLogs();
  });

  it('waits for logs even if trace flushing fails', async () => {
    const diagnostic = stub(diag, 'error');
    stub(NodeTracerProvider.prototype, 'forceFlush').rejects(
      new Error('trace flush failed'),
    );
    const { handler } = await loadHandler();
    assert.strictEqual(await handler({}, ctx), 'ok');
    assertLogs();
    assert(diagnostic.calledOnce);
  });

  it('waits for logs even if metric flushing fails', async () => {
    const diagnostic = stub(diag, 'error');
    stub(MeterProvider.prototype, 'forceFlush').rejects(
      new Error('metric flush failed'),
    );
    const { handler } = await loadHandler();
    assert.strictEqual(await handler({}, ctx), 'ok');
    assertLogs();
    assert(diagnostic.calledOnce);
  });

  it('preserves the result and flushes other signals if log flushing fails', async () => {
    const diagnostic = stub(diag, 'error');
    stub(LoggerProvider.prototype, 'forceFlush').rejects(
      new Error('log flush failed'),
    );
    const metricFlush = spy(MeterProvider.prototype, 'forceFlush');
    const { handler } = await loadHandler();
    assert.strictEqual(await handler({}, ctx), 'ok');
    assert(metricFlush.calledOnce);
    assert.strictEqual(spanExporter.getFinishedSpans().length, 1);
    assert(diagnostic.calledOnce);
  });

  it('keeps a separately registered custom logger provider', async () => {
    const customProvider = new LoggerProvider();
    const customFlush = spy(customProvider, 'forceFlush');
    global.configureLoggerProvider = provider => {
      loggerProvider = provider as LoggerProvider;
      logs.setGlobalLoggerProvider(customProvider);
    };
    await wrap();
    const { handler } = require('./handler/cjs/logs.js');
    try {
      assert.strictEqual(await handler({}, ctx), 'ok');
      assert.strictEqual(logs.getLoggerProvider(), customProvider);
      assert(customFlush.notCalled);
    } finally {
      await customProvider.shutdown();
    }
  });

  it('exports logs from each warm invocation and its response hook', async () => {
    global.configureLambdaInstrumentation = config => ({
      ...config,
      responseHook: () => logs.getLogger('test').emit({ body: 'response log' }),
    });
    const { handler } = await loadHandler();
    for (let i = 0; i < 2; i++) {
      assert.strictEqual(await handler({}, ctx), 'ok');
      assertLogs(['handler log', 'response log']);
      exporter.reset();
    }
  });
});
