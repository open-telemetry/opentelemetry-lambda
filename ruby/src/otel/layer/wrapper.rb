require 'opentelemetry-sdk'
require 'opentelemetry-exporter-otlp'
require 'opentelemetry-instrumentation-all'

# We need to load the function code's dependencies, and _before_ any dependencies might
# be initialized outside of the function handler, bootstrap instrumentation.
def preload_function_dependencies
  default_task_location = '/var/task'

  handler_file = ENV.values_at('ORIG_HANDLER', '_HANDLER').compact.first&.split('.')&.first

  unless handler_file && File.exist?("#{default_task_location}/#{handler_file}.rb")
    OpenTelemetry.logger.warn { 'Could not find the original handler file to preload libraries.' }
    return nil
  end

  libraries = File.read("#{default_task_location}/#{handler_file}.rb")
                  .scan(/^\s*require\s+['"]([^'"]+)['"]/)
                  .flatten

  libraries.each do |lib|
    require lib
  rescue StandardError => e
    OpenTelemetry.logger.warn { "Could not load library #{lib}: #{e.message}" }
  end
  handler_file
end

handler_file = preload_function_dependencies

OpenTelemetry.logger.info { "Libraries in #{handler_file} have been preloaded." } if handler_file

OpenTelemetry::SDK.configure do |c|
  c.use_all()
end

def otel_wrapper(event:, context:)
  otel_wrapper = OpenTelemetry::Instrumentation::AwsLambda::Handler.new
  otel_wrapper.call_wrapped(event: event, context: context)
end
