import path from 'path';
import url from 'url';

import assert from 'assert';

import {
  SpanKind,
  SpanStatusCode
} from '@opentelemetry/api';
import {
  BatchSpanProcessor,
  InMemorySpanExporter
} from '@opentelemetry/sdk-trace-node';

import { registerLoader } from '../src/loader.mjs';
import { init, wrap, unwrap } from '../build/src/wrapper.js';

const DIR_NAME = path.dirname(url.fileURLToPath(import.meta.url));

const assertHandlerSpan = (span) => {
  assert.strictEqual(span.kind, SpanKind.SERVER);
  assert.strictEqual(span.name, 'my_function');
  assert.strictEqual(span.attributes['faas.id'], 'my_arn');
  assert.strictEqual(span.status.code, SpanStatusCode.UNSET);
  assert.strictEqual(span.status.message, undefined);
};

describe('when loading ESM module', async () => {
  let oldEnv;
  const memoryExporter = new InMemorySpanExporter();

  const ctx = {
    functionName: 'my_function',
    invokedFunctionArn: 'my_arn',
    awsRequestId: 'aws_request_id',
  };

  await init();

  const initializeHandler = async (handler) => {
    process.env._HANDLER = handler;

    global.configureTracer = (_) => {
      return {
        spanProcessors: [new BatchSpanProcessor(memoryExporter)],
      };
    };
    global.configureMeter = (_) => { {} };
    global.configureMeterProvider = (_) => {};
    global.configureLoggerProvider = (_) => {};

    await wrap();
  };

  const loadHandler = async (handler) => {
    return await import(path.join(DIR_NAME, handler));
  };

  const initializeAndLoadHandler = async (handler, handlerFileName) => {
    await initializeHandler(handler);
    return await loadHandler(handlerFileName);
  };

  before(() => {
    registerLoader();
  });

  beforeEach(async () => {
    oldEnv = { ...process.env };
    process.env.LAMBDA_TASK_ROOT = DIR_NAME;

    await unwrap();
  });

  afterEach(async () => {
    process.env = oldEnv;
    memoryExporter.reset();

    await unwrap();
  });

  it('should wrap ESM file handler with .mjs extension', async () => {
    const lambdaModule = await initializeAndLoadHandler('./handler/esm/index_esm.handler', './handler/esm/index_esm.mjs');
    const result = await lambdaModule.handler('arg', ctx);

    assert.strictEqual(result, 'ok');

    const spans = memoryExporter.getFinishedSpans();
    assert.strictEqual(spans.length, 1);
    const [span] = spans;
    assertHandlerSpan(span);
  });

  it('should wrap ESM module handler with .js extension', async () => {
    const lambdaModule = await initializeAndLoadHandler('./handler/esm/index.handler', './handler/esm/index.js');
    const result = await lambdaModule.handler('arg', ctx);

    assert.strictEqual(result, 'ok');

    const spans = memoryExporter.getFinishedSpans();
    assert.strictEqual(spans.length, 1);
    const [span] = spans;
    assertHandlerSpan(span);
  });
});
