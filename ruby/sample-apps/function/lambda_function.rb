require 'json'

def lambda_handler(event:, context:)
  if defined?(::OpenTelemetry::SDK)
    { statusCode: 200, body: "Hello #{::OpenTelemetry::SDK::VERSION}" }
  else
    { statusCode: 200, body: "Missing OpenTelemetry" }
  end
end
