# Integration Tests

This test suite contains a simple setup to deploy lambda functions using the otel layers. These functions then use the aws-sdk library provided in the lamba runtime to make an sts call. We evaluate whether the expected telemetry was generated for this aws-sdk call.
The setup is very basic, it serves more as a smoke check than an "all-covering" test suite.
