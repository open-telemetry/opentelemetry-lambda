import { defineConfig } from 'vitest/config';

const TEST_TIMEOUT_MS = 120_000;
const HOOK_TIMEOUT_MS = 600_000; // Generous timeout for CDK deploy and destroy

export default defineConfig({
  test: {
    include: ['tests/**/*.test.ts'],
    globalSetup: ['./globalSetup.ts'],
    testTimeout: TEST_TIMEOUT_MS,
    hookTimeout: HOOK_TIMEOUT_MS,
  },
});
