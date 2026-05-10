import { describe, it, expect, inject } from "vitest";
import { LambdaClient, InvokeCommand } from "@aws-sdk/client-lambda";
import { waitForSpans as waitForLogs } from "../helpers/cloudwatch.js";

const lambdaClient = new LambdaClient({});

describe("Python Lambda layer", () => {
  it("produces STS spans", async () => {
    const functionName = inject("functionName");
    const logGroupName = inject("logGroupName");

    const startTime = Date.now();

    const response = await lambdaClient.send(
      new InvokeCommand({
        FunctionName: functionName,
        InvocationType: "RequestResponse",
        Payload: JSON.stringify({}),
      }),
    );
    expect(response.StatusCode).toBe(200);

    const payload = JSON.parse(new TextDecoder().decode(response.Payload));
    const body = JSON.parse(payload.body);
    expect(body.status).toBe("ok");
    expect(body.account).toBeDefined();

    const logEvents = await waitForLogs({
      logGroupName,
      filterPattern: '"otelcol.component.id" "debug" "exporter"',
      startTime,
    });
    const traceEvents = logEvents
      .map((event) => {
        if (!event.message) throw new Error("CloudWatch event missing message");
        return JSON.parse(event.message) as Record<string, unknown>;
      })
      .filter((span) => span["otelcol.signal"] === "traces");
    expect(traceEvents).toHaveLength(2); // debug exporter emits 1 summary and 1 detailed log event

    const detailedTraceEventMsg = traceEvents[1].msg as string;
    const instrumentationScopes = Array.from(
      detailedTraceEventMsg.matchAll(/^InstrumentationScope (\S+)/gm),
      (match) => match[1]
    );
    expect(instrumentationScopes).toEqual([
      "opentelemetry.instrumentation.botocore",
      "opentelemetry.instrumentation.aws_lambda",
    ]);
  });
});
