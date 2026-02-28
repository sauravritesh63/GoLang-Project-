import { test, expect } from "@playwright/test";

/**
 * Integration tests for the Next.js dashboard.
 *
 * These tests require the dev/production server to be running.
 * Run with: npx playwright test
 * (The playwright.config.ts webServer config starts `npm run dev` automatically
 * in non-CI environments.)
 *
 * In CI, set PLAYWRIGHT_BASE_URL to the deployed URL.
 */

test.describe("Dashboard page", () => {
  test.beforeEach(async ({ page }) => {
    await page.goto("/");
  });

  test("has correct page title", async ({ page }) => {
    await expect(page).toHaveTitle(/Task Scheduler Dashboard/i);
  });

  test("sidebar navigation links are visible", async ({ page }) => {
    await expect(page.getByRole("link", { name: /dashboard/i })).toBeVisible();
    await expect(page.getByRole("link", { name: /workflows/i })).toBeVisible();
    await expect(page.getByRole("link", { name: /workflow runs/i })).toBeVisible();
    await expect(page.getByRole("link", { name: /task runs/i })).toBeVisible();
    await expect(page.getByRole("link", { name: /workers/i })).toBeVisible();
  });

  test("stat cards section is rendered", async ({ page }) => {
    await expect(page.getByText(/total workflows/i)).toBeVisible();
    await expect(page.getByText(/active workers/i)).toBeVisible();
  });

  test("navigates to workflows page via sidebar", async ({ page }) => {
    await page.getByRole("link", { name: /^workflows$/i }).click();
    await expect(page).toHaveURL(/\/workflows/);
  });

  test("navigates to workers page via sidebar", async ({ page }) => {
    await page.getByRole("link", { name: /workers/i }).click();
    await expect(page).toHaveURL(/\/workers/);
  });
});

test.describe("Workflows page", () => {
  test("renders workflow list heading", async ({ page }) => {
    await page.goto("/workflows");
    await expect(page.getByText(/workflows/i).first()).toBeVisible();
  });
});

test.describe("Workers page", () => {
  test("renders workers heading", async ({ page }) => {
    await page.goto("/workers");
    await expect(page.getByText(/workers/i).first()).toBeVisible();
  });
});

test.describe("Workflow Runs page", () => {
  test("renders workflow runs content", async ({ page }) => {
    await page.goto("/workflow-runs");
    await expect(page.getByText(/workflow runs/i).first()).toBeVisible();
  });
});

test.describe("Task Runs page", () => {
  test("renders task runs content", async ({ page }) => {
    await page.goto("/task-runs");
    await expect(page.getByText(/task runs/i).first()).toBeVisible();
  });
});
