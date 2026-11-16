import { expect, test } from "@playwright/test";
import { signIn } from "./helpers";

// Epic 6: a learner sends feedback; the owner signs in to the admin board, finds and resolves it.
// scripts/e2e.sh seeds the admin account (ADMIN_USERNAME / ADMIN_PASSWORD on `api seed`).
const ADMIN_USERNAME = process.env.E2E_ADMIN_USERNAME ?? "owner";
const ADMIN_PASSWORD = process.env.E2E_ADMIN_PASSWORD ?? "e2e-admin-password";
const SHOTS = process.env.E2E_SCREENSHOTS;

test.beforeEach(({ page }) => {
  page.on("pageerror", (error) => console.log(`[pageerror] ${error.message}\n${error.stack ?? ""}`));
});

test("learner sends feedback, admin reads and resolves it", async ({ page, context, baseURL }, info) => {
  const shot = async (name: string) => {
    if (SHOTS) await page.screenshot({ path: `${SHOTS}/${info.project.name}-${name}.png`, fullPage: true });
  };
  const message = `The audio button does nothing on the study screen (${info.project.name} ${Date.now()})`;

  // Learner side.
  await signIn(context, baseURL!);
  await page.goto("/en/decks");
  await page.getByRole("button", { name: "Open account menu" }).click();
  await page.getByRole("menuitem", { name: "Send feedback" }).click();
  const dialog = page.getByRole("dialog");
  await expect(dialog.getByRole("heading", { name: "Send feedback" })).toBeVisible();
  await dialog.getByRole("radio", { name: "An idea" }).click();
  await dialog.getByLabel("Your message").fill("short");
  await expect(dialog.getByRole("button", { name: "Send" })).toBeDisabled();
  await dialog.getByLabel("Your message").fill(message);
  await shot("60-feedback-dialog");
  await dialog.getByRole("button", { name: "Send" }).click();
  await expect(page.getByText("Thanks! We read every message.")).toBeVisible();
  await expect(dialog).toBeHidden();

  // The board is closed without the admin session: the learner's token is not an admin token.
  await page.goto("/en/admin/feedback");
  await expect(page).toHaveURL(/\/en\/admin\/login\?next=/);
  const direct = await page.request.get("/api/v1/admin/overview");
  expect(direct.status()).toBe(401);

  // A wrong password is rejected.
  await page.getByLabel("Username").fill(ADMIN_USERNAME);
  await page.getByLabel("Password").fill("not-the-password");
  await page.getByRole("button", { name: "Sign in" }).click();
  await expect(page.getByText("Wrong username or password.")).toBeVisible();

  await page.getByLabel("Password").fill(ADMIN_PASSWORD);
  await page.getByRole("button", { name: "Sign in" }).click();
  await expect(page).toHaveURL(/\/en\/admin\/feedback$/);

  // Feedback board: the new message with its context.
  const item = page.getByRole("article").filter({ hasText: message });
  await expect(item).toBeVisible();
  await expect(item.getByText("An idea")).toBeVisible();
  await expect(item.getByText("/en/decks")).toBeVisible();
  await shot("61-admin-feedback");
  await item.getByRole("button", { name: "Add note" }).click();
  await item.getByLabel("Private note").fill("Check iOS Safari speech");
  await item.getByRole("button", { name: "Save note" }).click();
  await expect(page.getByText("Note saved")).toBeVisible();
  await page.getByRole("article").filter({ hasText: message }).getByRole("button", { name: "Resolve" }).click();
  await expect(page.getByRole("article").filter({ hasText: message })).toBeHidden();
  await page.getByRole("tab", { name: /Resolved/ }).click();
  const resolved = page.getByRole("article").filter({ hasText: message });
  await expect(resolved.getByText("Check iOS Safari speech")).toBeVisible();
  await expect(resolved.getByRole("button", { name: "Reopen" })).toBeVisible();

  // Overview and users (tabs in the admin navigation; stat tiles are links too).
  const tabs = page.getByRole("navigation", { name: "Admin sections" });
  await tabs.getByRole("link", { name: "Overview" }).click();
  await expect(page.getByRole("heading", { name: "Overview" })).toBeVisible();
  await expect(page.getByRole("heading", { name: "Last 30 days" })).toBeVisible();
  await shot("62-admin-overview");
  await tabs.getByRole("link", { name: "Users" }).click();
  await page.getByLabel("Search by email or name").fill("e2e-");
  await expect(page.getByRole("cell", { name: /Minh Anh/ }).first()).toBeVisible();
  await shot("63-admin-users");

  // Sign out closes the board again.
  await page.getByRole("button", { name: "Sign out" }).click();
  await expect(page).toHaveURL(/\/en\/admin\/login/);
  await page.goto("/en/admin");
  await expect(page).toHaveURL(/\/en\/admin\/login/);
});
