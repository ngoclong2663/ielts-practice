import { defineConfig, devices } from "@playwright/test";

// End-to-end tests run against a test-mode API (APP_ENV=test, which enables /auth/test-login)
// and a production build of the web app. See the README "End-to-end tests" section.
export default defineConfig({
  testDir: "./e2e",
  timeout: 60_000,
  fullyParallel: false,
  retries: 0,
  reporter: [["list"]],
  use: {
    baseURL: process.env.E2E_BASE_URL ?? "http://localhost:3001",
    trace: "retain-on-failure",
  },
  projects: [
    { name: "desktop", use: { ...devices["Desktop Chrome"] } },
    { name: "mobile", use: { ...devices["Pixel 7"] } },
  ],
});
