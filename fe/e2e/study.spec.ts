import { expect, test } from "@playwright/test";
import { signIn } from "./helpers";

// Surface client-side crashes in the test output.
test.beforeEach(({ page }) => {
  page.on("pageerror", (error) => console.log(`[pageerror] ${error.message}\n${error.stack ?? ""}`));
});

const SHOTS = process.env.E2E_SCREENSHOTS;

test("create a deck, add a card and study it", async ({ page, context, baseURL }, info) => {
  await signIn(context, baseURL!);
  const shot = async (name: string) => {
    if (SHOTS) await page.screenshot({ path: `${SHOTS}/${info.project.name}-${name}.png`, fullPage: true });
  };

  // Empty dashboard points to decks.
  await page.goto("/en/dashboard");
  await expect(page.getByRole("heading", { name: "No cards to review yet" })).toBeVisible();
  await shot("01-dashboard-empty");

  // Create a deck.
  await page.goto("/en/decks");
  await page.getByRole("button", { name: "New deck" }).first().click();
  await page.getByLabel("Name").fill("IELTS Environment");
  await page.getByRole("button", { name: "Create deck" }).click();
  await expect(page.getByText("IELTS Environment")).toBeVisible();

  // Add two cards.
  await page.getByRole("link", { name: /IELTS Environment/ }).click();
  for (const [word, meaning] of [
    ["sustainable", "bền vững"],
    ["pollution", "sự ô nhiễm"],
  ]) {
    await page.getByRole("button", { name: "Add card" }).click();
    await page.getByLabel("Word or phrase").fill(word);
    await page.getByLabel("Vietnamese meaning").fill(meaning);
    await page.getByLabel(/Example sentence/).fill(`Cities need ${word} transport.`);
    await page.getByRole("button", { name: "Save card" }).click();
    await expect(page.getByRole("button", { name: word })).toBeVisible();
  }
  // Duplicate warning.
  await page.getByRole("button", { name: "Add card" }).click();
  await page.getByLabel("Word or phrase").fill("Pollution");
  await page.getByLabel("Vietnamese meaning").fill("x");
  await page.getByRole("button", { name: "Save card" }).click();
  await expect(page.getByText("This word is already in the deck")).toBeVisible();
  await page.getByRole("button", { name: "Cancel" }).click();
  await shot("02-deck-detail");

  // Study: reveal with Space, answer with the keyboard and the buttons.
  await page.getByRole("link", { name: "Study", exact: true }).click();
  await expect(page.getByRole("heading", { name: "sustainable" })).toBeVisible();
  await shot("03-study-front");
  await page.keyboard.press("Space");
  // Recently visited pages stay mounted but hidden (React Activity), so only match visible text.
  await expect(page.getByText("bền vững").filter({ visible: true })).toBeVisible();
  await expect(page.getByRole("button", { name: /1m.*Again/ })).toBeVisible();
  await shot("04-study-back");
  await page.keyboard.press("4"); // Easy: graduates the first card

  await expect(page.getByRole("heading", { name: "pollution" })).toBeVisible();
  await page.getByRole("button", { name: "Show answer" }).click();
  await page.getByRole("button", { name: /Good/ }).click(); // learning, back in 10 minutes

  // Nothing else is left, so the learning card comes back early (learn-ahead).
  await expect(page.getByRole("heading", { name: "pollution" })).toBeVisible();
  await page.getByRole("button", { name: "Undo" }).click();
  await expect(page.getByText("Last answer undone")).toBeVisible();
  await page.getByRole("button", { name: "Show answer" }).click();
  await page.getByRole("button", { name: /Easy/ }).click();

  await expect(page.getByRole("heading", { name: "You're done for today" })).toBeVisible();
  await expect(page.getByText("Cards studied")).toBeVisible();
  await shot("05-study-done");

  // Dashboard reflects the session.
  await page.goto("/en/dashboard");
  await expect(page.getByText("Reviewed today")).toBeVisible();
  await expect(page.getByText("1 day streak")).toBeVisible();
  await shot("06-dashboard");
});

test("the deck pages work in Vietnamese", async ({ page, context, baseURL }, info) => {
  await signIn(context, baseURL!, "vi");
  await page.goto("/vi/decks");
  await expect(page.getByRole("heading", { name: "Bộ thẻ của tôi" })).toBeVisible();
  await expect(page.getByText("Chưa có bộ thẻ nào")).toBeVisible();
  if (SHOTS) await page.screenshot({ path: `${SHOTS}/${info.project.name}-07-decks-vi.png`, fullPage: true });
});

test("add a card from the Add words page", async ({ page, context, baseURL }, info) => {
  await signIn(context, baseURL!);
  await page.goto("/en/add-words");
  await page.getByRole("tab", { name: "Add one by hand" }).click();
  await expect(page.getByText("Create a deck first")).toBeVisible();
  await page.getByRole("button", { name: "New deck" }).click();
  await page.getByLabel("Name").fill("Quick words");
  await page.getByRole("button", { name: "Create deck" }).click();

  await page.getByRole("button", { name: "Add card" }).click();
  await page.getByLabel("Word or phrase").fill("resilient");
  await page.getByLabel("Vietnamese meaning").fill("kiên cường");
  await page.getByRole("button", { name: "Save card" }).click();
  await expect(page.getByRole("link", { name: /1 card added/ })).toBeVisible();
  if (SHOTS) await page.screenshot({ path: `${SHOTS}/${info.project.name}-08-add-words.png`, fullPage: true });

  await page.getByRole("link", { name: /1 card added/ }).click();
  await expect(page.getByRole("button", { name: "resilient" })).toBeVisible();
});
