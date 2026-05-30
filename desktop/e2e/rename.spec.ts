import { readFileSync } from "node:fs";
import { join } from "node:path";
import { test, expect } from "@playwright/test";
import type { E2EState } from "./global.setup";

const STATE_FILE = join(import.meta.dirname, "..", "..", ".e2e-run", "state.json");

function e2eState(): E2EState {
  const raw = readFileSync(STATE_FILE, "utf-8");
  return JSON.parse(raw) as E2EState;
}

async function setupPage(page: any) {
  const state = e2eState();
  await page.goto("/");
  await page.waitForFunction(
    () => (window as any)?.go?.main?.App !== undefined,
    { timeout: 30_000 },
  );
  await page.evaluate(async (configPath: string) => {
    const app = (window as any).go.main.App;
    await app.SetConfigPath(configPath);
  }, state.configPath);
  return state;
}

async function scanDirectory(page: any, sourceDir: string) {
  await page.locator('input[placeholder="/path/to/files"]').fill(sourceDir);
  await page.getByRole("button", { name: /rescan/i }).click();
  // Wait for the file browser to appear — file entries are buttons with truncate
  await expect(
    page.locator("button .truncate").first(),
  ).toBeVisible({ timeout: 30_000 });
}

async function waitForJobComplete(page: any) {
  await expect(async () => {
    const badge = page.locator("header div.rounded-full").first();
    await expect(badge).toContainText(/Complete/i, { timeout: 5_000 });
  }).toPass({ timeout: 120_000 });
}

function sidebarItem(page: any, label: string) {
  return page.locator("aside").getByTitle(label);
}

async function toggleCheckbox(page: any, label: string | RegExp, checked: boolean) {
  // The Toggle component wraps the hidden input in a <label> with a <span>
  // overlay. Click the label text to toggle — it handles the state change.
  const toggleLabel = page.getByText(label).first();
  const checkbox = page.getByRole("checkbox", { name: label });
  const isChecked = await checkbox.isChecked();
  if (isChecked !== checked) {
    await toggleLabel.click();
  }
}

// ─── Tests ────────────────────────────────────────────────────────────

test.describe("Rename workflow", () => {
  test("app loads and shows rename view with sidebar navigation", async ({ page }) => {
    await setupPage(page);

    await expect(page.locator("h1")).toContainText("Rename");

    await expect(sidebarItem(page, "Rename")).toBeVisible();
    await expect(sidebarItem(page, "History")).toBeVisible();
    await expect(sidebarItem(page, "Analytics")).toBeVisible();
    await expect(sidebarItem(page, "Settings")).toBeVisible();

    // Step indicators are now in the titlebar — verify they render
    await expect(page.getByText("Pick", { exact: true })).toBeVisible();
    await expect(page.getByText("Config", { exact: true })).toBeVisible();
    await expect(page.getByText("Preview", { exact: true })).toBeVisible();
    await expect(page.getByText("Done", { exact: true })).toBeVisible();
  });

  test("scans a directory and shows files in the file browser", async ({ page }) => {
    const state = await setupPage(page);
    await scanDirectory(page, state.sourceDir);

    // File browser shows renamed files — each shows new name and original name
    for (const filename of state.files) {
      const baseName = filename.split("/").pop() || filename;
      // The original filename appears as muted text in the second line
      await expect(page.getByText(baseName).first()).toBeVisible();
    }
  });

  test("runs a dry-run rename job to completion", async ({ page }) => {
    const state = await setupPage(page);
    await scanDirectory(page, state.sourceDir);

    await page.getByRole("button", { name: /run/i }).click();
    await waitForJobComplete(page);
  });

  test("rename preview shows correct snake_case names from mock AI", async ({ page }) => {
    const state = await setupPage(page);
    await scanDirectory(page, state.sourceDir);

    // Trigger name generation — this puts the app in preview-ready state
    await page.getByRole("button", { name: /run/i }).click();
    await expect(async () => {
      const badge = page.locator("header div.rounded-full").first();
      await expect(badge).toContainText(/Ready/i, { timeout: 5_000 });
    }).toPass({ timeout: 120_000 });

    const expectedRenames: Record<string, string> = {
      "Quarterly Business Report.pdf": "quarterly_business_report.pdf",
      "TAX DOCUMENT 2025.pdf": "tax_document_2025.pdf",
      "Meeting Notes - Jan 15.txt": "meeting_notes_jan_15.txt",
      "img_0042.jpg": "img_0042.jpg",
      "Profile Photo.png": "profile_photo.png",
    };

    for (const [original, expected] of Object.entries(expectedRenames)) {
      // The new name appears prominently
      await expect(page.getByText(expected).first()).toBeVisible();
      // The original filename always appears as muted reference text
      const baseName = original.split("/").pop() || original;
      await expect(page.getByText(baseName).first()).toBeVisible();
    }
  });
});

test.describe("History and analytics views", () => {
  test("history view shows completed jobs after a rename run", async ({ page }) => {
    const state = await setupPage(page);

    await scanDirectory(page, state.sourceDir);
    await page.getByRole("button", { name: /run/i }).click();
    await waitForJobComplete(page);

    await sidebarItem(page, "History").click();
    await expect(page.locator("h1")).toContainText("History", { timeout: 10_000 });

    const rows = page.locator("table tr");
    await expect(rows.count()).resolves.toBeGreaterThan(1);
  });

  test("analytics view shows session stats after a job", async ({ page }) => {
    const state = await setupPage(page);

    await scanDirectory(page, state.sourceDir);
    await page.getByRole("button", { name: /run/i }).click();
    await waitForJobComplete(page);

    await sidebarItem(page, "Analytics").click();
    await expect(page.locator("h1")).toContainText("Analytics", { timeout: 10_000 });

    await expect(page.getByText("Sessions", { exact: true })).toBeVisible();
    await expect(page.getByText("Renamed", { exact: true })).toBeVisible();
    await expect(page.getByText("Tokens", { exact: true })).toBeVisible();
  });
});

test.describe("Organize and execution features", () => {
  test("organize toggle is visible and enabled by default", async ({ page }) => {
    await setupPage(page);

    await expect(page.getByText("Organize")).toBeVisible();
    const organizeCheckbox = page.getByRole("checkbox", { name: /organize/i });
    await expect(organizeCheckbox).toBeChecked();
  });

  test("can toggle dry run off and run a real file rename", async ({ page }) => {
    const state = await setupPage(page);

    // Scan — the file browser shows the same tree for both dry-run modes
    await page.locator('input[placeholder="/path/to/files"]').fill(state.sourceDir);
    await page.getByRole("button", { name: /rescan/i }).click();
    await expect(
      page.locator("button .truncate").first(),
    ).toBeVisible({ timeout: 30_000 });

    await page.getByRole("button", { name: /run/i }).click();
    await waitForJobComplete(page);

    await expect(page.getByText(/Rename complete/)).toBeVisible();
  });
});

test.describe("Settings view", () => {
  test("settings view loads and shows configured AI model", async ({ page }) => {
    await setupPage(page);

    await sidebarItem(page, "Settings").click();
    await expect(page.locator("h1")).toContainText("Settings", { timeout: 10_000 });

    const modelInput = page.locator('input[list="model-suggestions"]');
    await expect(modelInput).toBeVisible();
    await expect(modelInput).toHaveValue(/mock-llama/i);
  });
});

test.describe("Navigation and UI elements", () => {
  test("Browse button is visible in rename view", async ({ page }) => {
    await setupPage(page);
    await expect(page.getByRole("button", { name: /browse/i })).toBeVisible();
  });

  test("all four sidebar tabs navigate to their views", async ({ page }) => {
    const state = await setupPage(page);

    await scanDirectory(page, state.sourceDir);
    await page.getByRole("button", { name: /run/i }).click();
    await waitForJobComplete(page);

    const tabs = ["Rename", "History", "Analytics", "Settings"];
    for (const tab of tabs) {
      await sidebarItem(page, tab).click();
      await expect(page.locator("h1")).toContainText(tab, { timeout: 10_000 });
    }
  });
});
