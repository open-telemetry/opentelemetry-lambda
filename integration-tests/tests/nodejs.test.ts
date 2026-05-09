import { describe, it, expect, inject } from 'vitest';
import { LambdaClient, InvokeCommand } from '@aws-sdk/client-lambda';
import { waitForSpans } from '../helpers/cloudwatch.js';

const lambdaClient = new LambdaClient({});

describe('Node.js Lambda layer', () => {
  it('produces STS spans via the debug exporter', async () => {
    const functionName = inject('functionName');
    const logGroupName = inject('logGroupName');

    const startTime = Date.now();

    const response = await lambdaClient.send(
      new InvokeCommand({
        FunctionName: functionName,
        InvocationType: 'RequestResponse',
        Payload: JSON.stringify({}),
      }),
    );

    const payload = JSON.parse(new TextDecoder().decode(response.Payload));
    expect(response.StatusCode).toBe(200);
    expect(payload.statusCode).toBe(200);

    const body = JSON.parse(payload.body);
    expect(body.status).toBe('ok');
    expect(body.account).toBeDefined();

    const events = await waitForSpans({
      logGroupName,
      filterPattern: '"STS" "GetCallerIdentity"',
      startTime,
      timeoutMs: 60_000,
    });

    expect(events.length).toBeGreaterThan(0);
  });
});
