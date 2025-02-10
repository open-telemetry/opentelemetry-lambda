const WRAPPER_INIT_START_TIME = Date.now();
const { default: wrapper } = await import('./wrapper.js');
await wrapper.init();
await wrapper.wrap();
console.log('OpenTelemetry wrapper init completed in', Date.now() - WRAPPER_INIT_START_TIME, 'ms');

const LOADER_INIT_START_TIME = Date.now();
await import('./loader.mjs');
console.log('OpenTelemetry loader init completed in', Date.now() - LOADER_INIT_START_TIME, 'ms');
