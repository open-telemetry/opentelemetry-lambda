# Integration Tests

This test suite contains a simple setup to deploy lambda functions using the otel layers. These functions then use the aws-sdk library provided in the lamba runtime to make an sts call. We evaluate whether the expected telemetry was generated for this aws-sdk call.
The setup is very basic, it serves more as a smoke check than an "all-covering" test suite.

## How it works

A single parameterized test (`tests/layer.test.ts`) runs for one language per invocation, selected via the `TEST_LANGUAGE` env var. Each language's runtime, handler and expected instrumentation scopes live in the `LANGUAGE_CONFIG` map in `globalSetup.ts`, which deploys a CDK stack (real AWS resources, so AWS credentials are required) and provides the function/log-group/scopes to the test. To add a language, add an entry to `LANGUAGE_CONFIG` and a handler under `handlers/`.

Supported languages: `nodejs`, `python`, `ruby`, `javaagent`, `javawrapper`.

In CI, the test runs automatically on every layer release (gated before publishing) and can be triggered manually via the "Integration Tests" workflow.

## Prerequisites (GitHub Actions setup)

The tests run in CI under an AWS account that needs a one-time setup.

1. **CDK bootstrap** the account in `us-east-1` once, so the conventional CDK asset bucket (`cdk-*-assets-<account>-us-east-1`) exists. The stack deploys with `CliCredentialsStackSynthesizer`, which uses the caller's credentials directly but still uploads the lambda/layer assets to that bucket:

   ```
   cdk bootstrap aws://<account>/us-east-1
   ```

2. **Deploy the IAM stack** at `utils/aws-cloudformation/aws-cf-stack-for-integration-tests.yml`. It creates the GitHub OIDC provider and the `github-otel-lambda-integration-test-role` that GitHub Actions assumes (scoped to deploying `IntegrationTest-*` stacks and reading their logs in `us-east-1`):

   ```
   aws cloudformation deploy \
     --template-file utils/aws-cloudformation/aws-cf-stack-for-integration-tests.yml \
     --stack-name otel-lambda-integration-test-iam \
     --parameter-overrides GitHubOrgName=<org> RepositoryName=opentelemetry-lambda \
     --capabilities CAPABILITY_NAMED_IAM \
     --region us-east-1
   ```

   > If the account already has a `token.actions.githubusercontent.com` OIDC provider, drop the `GithubOIDC` resource from the template and point the role's trust policy at the existing one — only one OIDC provider per URL is allowed per account. The other `utils/aws-cloudformation-aws-cf-stack-for-layer-publish.yml` sets this up as well.

3. Set the stack's `RoleARN` output as the repository secret **`INTEGRATION_TEST_ROLE_ARN`** (consumed by `.github/workflows/integration-test.yml`).

## Running tests locally

Locally you only need AWS credentials configured (profile or SSO) with permissions to deploy the stack, in a `us-east-1` account that has been CDK-bootstrapped (see step 1 from prerequisites above).

Build the layer zips you want to test (collector + the language's instrumentation layer), then run from this directory:

```
TEST_LANGUAGE=nodejs \
COLLECTOR_LAYER_ZIP=/path/to/opentelemetry-collector-layer-amd64.zip \
INSTRUMENTATION_LAYER_ZIP=/path/to/opentelemetry-nodejs-layer.zip \
npm test
```

## Running the Java tests locally

The Java tests (`javaagent`, `javawrapper`) deploy a prebuilt handler jar. Build it first before running those tests locally:

```
cd handlers/java && ./gradlew shadowJar
```
