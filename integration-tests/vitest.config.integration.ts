import { defineConfig } from 'vitest/config';

export default defineConfig({
  test: {
    include: ['tests/**/*.test.ts'],
    globalSetup: ['./globalSetup.ts'],
    testTimeout: 120_000,
    hookTimeout: 600_000,
  },
});
