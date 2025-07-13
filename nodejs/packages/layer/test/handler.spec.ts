import path from 'path';

import assert from 'assert';

import { SpanKind, SpanStatusCode } from '@opentelemetry/api';
import {
  BatchSpanProcessor,
  InMemorySpanExporter,
  ReadableSpan,
} from '@opentelemetry/sdk-trace-node';
import { Context } from 'aws-lambda';

import { init, wrap, unwrap } from '../src/wrapper';

const assertHandlerSpan = (span: ReadableSpan) => {
  assert.strictEqual(span.kind, SpanKind.SERVER);
  assert.strictEqual(span.name, 'my_function');
  assert.strictEqual(span.attributes['faas.id'], 'my_arn');
  assert.strictEqual(span.status.code, SpanStatusCode.UNSET);
  assert.strictEqual(span.status.message, undefined);
};

describe('when loading ESM module', async () => {
  let oldEnv: NodeJS.ProcessEnv;
  const memoryExporter = new InMemorySpanExporter();

  const ctx = {
    functionName: 'my_function',
    invokedFunctionArn: 'my_arn',
    awsRequestId: 'aws_request_id',
  } as Context;

  await init();

  const initializeHandler = async (handler: string) => {
    process.env._HANDLER = handler;

    global.configureTracer = _ => {
      return {
        spanProcessors: [new BatchSpanProcessor(memoryExporter)],
      };
    };
    global.configureMeter = _ => {
      return {} as any;
    };
    global.configureMeterProvider = _ => {};
    global.configureLoggerProvider = _ => {};

    await wrap();
  };

  const loadHandler = async (handler: string) => {
    return await import(path.join(__dirname, handler));
  };

  const initializeAndLoadHandler = async (
    handler: string,
    handlerFileName: string,
  ) => {
    await initializeHandler(handler);
    return await loadHandler(handlerFileName);
  };

  beforeEach(async () => {
    oldEnv = { ...process.env };
    process.env.LAMBDA_TASK_ROOT = __dirname;

    await unwrap();
  });

  afterEach(async () => {
    process.env = oldEnv;
    memoryExporter.reset();

    await unwrap();
  });

  it('should wrap CommonJS file handler with .cjs extension', async () => {
    const lambdaModule = await initializeAndLoadHandler(
      './handler/cjs/index_commonjs.handler',
      './handler/cjs/index_commonjs.cjs',
    );
    const result = await lambdaModule.handler('arg', ctx);

    assert.strictEqual(result, 'ok');

    const spans = memoryExporter.getFinishedSpans();
    assert.strictEqual(spans.length, 1);
    const [span] = spans;
    assertHandlerSpan(span);
  });

  it('should wrap CommonJS module handler with .js extension', async () => {
    const lambdaModule = await initializeAndLoadHandler(
      './handler/cjs/index.handler',
      './handler/cjs/index.js',
    );
    const result = await lambdaModule.handler('arg', ctx);

    assert.strictEqual(result, 'ok');

    const spans = memoryExporter.getFinishedSpans();
    assert.strictEqual(spans.length, 1);
    const [span] = spans;
    assertHandlerSpan(span);
  });
});
