const { logs } = require('@opentelemetry/api-logs');

module.exports.handler = function (event, _context, callback) {
  logs.getLogger('test').emit({ body: 'handler log' });
  if (event.callback) {
    setImmediate(() => {
      logs.getLogger('test').emit({ body: 'callback log' });
      callback(event.error, 'ok');
    });
    return;
  }
  return event.error ? Promise.reject(event.error) : Promise.resolve('ok');
};

module.exports.streamingHandler = async function (event) {
  logs.getLogger('test').emit({ body: 'handler log' });
  if (event.error) {
    throw event.error;
  }
};
module.exports.streamingHandler[
  Symbol.for('aws.lambda.runtime.handler.streaming')
] = 'response';
