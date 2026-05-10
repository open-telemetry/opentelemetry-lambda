import { defineConfig } from 'vitest/config';

const TEST_TIMEOUT_MS = 120_000;
const HOOK_TIMEOUT_MS = 600_000; // Generous timeout for CDK deploy and destroy

const language = process.env.TEST_LANGUAGE;

export default defineConfig({
  test: {
    include: [language ? `tests/${language}.test.ts` : 'tests/**/*.test.ts'],
    globalSetup: ['./globalSetup.ts'],
    testTimeout: TEST_TIMEOUT_MS,
    hookTimeout: HOOK_TIMEOUT_MS,
  },
});
