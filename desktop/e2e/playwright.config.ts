import { defineConfig } from "@playwright/test";
import { join } from "node:path";

const WAILS_DEV_URL = process.env.WAILS_DEV_URL || "http://localhost:34115";
const CI = Boolean(process.env.CI);

// Keep Playwright output OUTSIDE the Wails project directory.
// Wails dev watches the project tree and triggers rebuilds when
// test artifacts are created/deleted, which kills the dev server.
const OUTPUT_DIR = join(import.meta.dirname, "..", "..", ".e2e-output");

export default defineConfig({
  testDir: ".",
  testMatch: "*.spec.ts",
  timeout: 60_000,
  expect: { timeout: 15_000 },

  globalSetup: "./global.setup.ts",
  globalTeardown: "./global.teardown.ts",

  outputDir: join(OUTPUT_DIR, "artifacts"),

  use: {
    baseURL: WAILS_DEV_URL,
    screenshot: "only-on-failure",
    video: CI ? "retain-on-failure" : "off",
    trace: CI ? "retain-on-failure" : "off",
  },

  projects: [
    {
      name: "chromium",
      use: {
        browserName: "chromium",
        launchOptions: {
          headless: true,
          args: ["--no-sandbox", "--disable-setuid-sandbox"],
        },
      },
    },
  ],

  ...(CI
    ? {
        workers: 1,
        retries: 1,
      }
    : {}),
});
