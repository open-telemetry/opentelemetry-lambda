import { AsyncLocalStorageContextManager } from '@opentelemetry/context-async-hooks';
import {
  BasicTracerProvider,
  PROPAGATOR_FACTORY,
  SDKRegistrationConfig,
  TracerConfig,
} from '@opentelemetry/sdk-trace-base';

export class LambdaTracerProvider extends BasicTracerProvider {
  protected static override readonly _registeredPropagators = new Map<
    string,
    PROPAGATOR_FACTORY
  >([...BasicTracerProvider._registeredPropagators]);
  constructor(config: TracerConfig = {}) {
    super(config);
  }
  override register(config: SDKRegistrationConfig = {}): void {
    if (config.contextManager === undefined) {
      config.contextManager = new AsyncLocalStorageContextManager();
      config.contextManager.enable();
    }
    super.register(config);
  }
}
