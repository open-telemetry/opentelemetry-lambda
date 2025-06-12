const WRAPPER_INIT_START_TIME = Date.now();
const { default: wrapper } = await import('./wrapper.js');
await wrapper.init();
await wrapper.wrap();

wrapper.logDebug('OpenTelemetry wrapper init completed in', Date.now() - WRAPPER_INIT_START_TIME, 'ms');

const LOADER_INIT_START_TIME = Date.now();
await import('./loader.mjs');
wrapper.logDebug('OpenTelemetry loader init completed in', Date.now() - LOADER_INIT_START_TIME, 'ms');
