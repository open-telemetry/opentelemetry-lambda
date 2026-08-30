import { WinstonInstrumentation } from '@opentelemetry/instrumentation-winston';
import type * as Winston from 'winston';
import assert from 'assert';

describe('winston instrumentation', () => {
  it('auto-attaches the OpenTelemetry transport so logs are exported', () => {
    // The instrumentation must be enabled (installing its require hook) before
    // 'winston' is required for the first time, otherwise the module load isn't
    // intercepted and patched - matching how createInstrumentations() in
    // wrapper.ts constructs instrumentations before the handler module (and its
    // winston usage) is loaded.
    const instrumentation = new WinstonInstrumentation();

    try {
      // Its patched `configure()` requires '@opentelemetry/winston-transport' to
      // attach an OTLP-exporting transport; if that package is not bundled with
      // the layer, the require fails and winston logs are silently never exported.
      const winston: typeof Winston = require('winston');

      const logger = winston.createLogger({
        level: 'info',
        transports: [new winston.transports.Console()],
      });

      const transportNames = logger.transports.map(t => t.constructor.name);
      assert.ok(
        transportNames.includes('OpenTelemetryTransportV3'),
        `expected an OpenTelemetryTransportV3 transport to be attached, got: ${transportNames}`,
      );
    } finally {
      instrumentation.disable();
    }
  });
});
