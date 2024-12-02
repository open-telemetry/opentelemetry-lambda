import type { AwsSdkInstrumentationConfig } from '@opentelemetry/instrumentation-aws-sdk';
import { stub } from 'sinon';

import { wrap, unwrap } from '../src/wrapper';

declare global {
  function configureAwsInstrumentation(
    defaultConfig: AwsSdkInstrumentationConfig,
  ): AwsSdkInstrumentationConfig;
}

const assert = require('assert');

describe('wrapper', () => {
  beforeEach(() => {
    unwrap();
  });

  afterEach(() => {
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
});
