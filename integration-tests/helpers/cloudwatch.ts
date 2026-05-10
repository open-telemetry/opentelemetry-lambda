import { setTimeout as sleep } from 'node:timers/promises';
import {
  CloudWatchLogsClient,
  FilterLogEventsCommand,
  type FilteredLogEvent,
} from '@aws-sdk/client-cloudwatch-logs';

const cwl = new CloudWatchLogsClient({});

export async function waitForSpans(options: {
  logGroupName: string;
  filterPattern: string;
  startTime: number;
  timeoutMs?: number;
  pollIntervalMs?: number;
}): Promise<FilteredLogEvent[]> {
  const {
    logGroupName,
    filterPattern,
    startTime,
    timeoutMs = 60_000,
    pollIntervalMs = 5_000,
  } = options;

  const deadline = Date.now() + timeoutMs;

  while (Date.now() < deadline) {
    const response = await cwl.send(
      new FilterLogEventsCommand({
        logGroupName,
        filterPattern,
        startTime,
      }),
    );

    if (response.events && response.events.length > 0) {
      return response.events;
    }

    await sleep(pollIntervalMs);
  }

  throw new Error(
    `Timed out waiting for spans matching "${filterPattern}" in ${logGroupName} after ${timeoutMs}ms`,
  );
}
