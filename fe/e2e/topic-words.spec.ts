import { expect, test } from "@playwright/test";
import { signIn } from "./helpers";

// Epic 4: topic → suggested words (no duplicates in the deck) → review → Epic 3 cards → deck.
// The e2e API uses VOCAB_FAKE_PROVIDERS=true: words come from a fixed pool; "diagnosis" is always
// proposed (like a model ignoring the exclude list), so the server-side de-duplication is visible.

test.beforeEach(({ page }) => {
  page.on("pageerror", (error) => console.log(`[pageerror] ${error.message}\n${error.stack ?? ""}`));
});

const SHOTS = process.env.E2E_SCREENSHOTS;

test("suggest words for a topic, review them and save cards into a deck", async ({ page, context, baseURL }, info) => {
  await signIn(context, baseURL!);
  const shot = async (name: string) => {
    if (SHOTS) await page.screenshot({ path: `${SHOTS}/${info.project.name}-${name}.png`, fullPage: true });
  };

  await page.goto("/en/add-words");
  await page.getByRole("tab", { name: "By topic" }).click();
  // Wait for the topic form (the paste tab has its own "create a deck" button).
  await expect(page.getByRole("button", { name: "Health & hospital" })).toBeVisible();
  await page.getByRole("button", { name: "or create a new deck" }).click();
  await page.getByLabel("Name").fill("Hospital");
  await page.getByRole("button", { name: "Create deck" }).click();
  await expect(page.getByRole("combobox", { name: "Save to" })).toHaveText(/Hospital/);

  await page.getByRole("button", { name: "Health & hospital" }).click();
  await page.getByRole("button", { name: "Fewer words" }).click();
  await page.getByRole("button", { name: "Fewer words" }).click();
  await expect(page.getByText("5 of 5 word lists left today.")).toBeVisible();
  await shot("30-topic-form");
  await page.getByRole("button", { name: "Suggest 5 words" }).click();

  await expect(page).toHaveURL(/suggestion=/);
  await expect(page.getByRole("heading", { name: "Words for “Health & hospital”" })).toBeVisible();
  await expect(page.getByText("5 words")).toBeVisible();

  // Untick one word; "I know this" hides another for good.
  await page.getByRole("checkbox", { name: "Keep symptom" }).click();
  const surgery = page.getByRole("listitem").filter({ hasText: "surgery" });
  await surgery.getByRole("button", { name: "I know this" }).click();
  await expect(page.getByText("“surgery” won't be suggested again")).toBeVisible();
  await expect(surgery).toHaveCount(0);

  // "Suggest more" adds new words only.
  await page.getByRole("button", { name: "Suggest more" }).click();
  await expect(page.getByText("therapy", { exact: true })).toBeVisible();
  await expect(page.getByText("3 word lists left today").locator("visible=true")).toBeVisible();
  await shot("31-word-list");

  // 9 words: 5 + 5 more, minus "surgery" (known); "symptom" is unticked → 8 cards.
  await page.getByRole("button", { name: "Generate 8 cards" }).click();
  await expect(page).toHaveURL(/job=/);
  await expect(page.getByText("8 of 8 cards ready")).toBeVisible({ timeout: 15_000 });
  await expect(page.getByRole("combobox", { name: "Save to" })).toHaveText(/Hospital/);
  await page.getByRole("button", { name: "Save 8 cards" }).click();
  await expect(page.getByText("Saved to Hospital")).toBeVisible();

  // Next time, words already in the deck are skipped.
  await page.getByRole("button", { name: "New list" }).click();
  await page.getByRole("tab", { name: "By topic" }).click();
  await page.getByRole("button", { name: "Health & hospital" }).click();
  await page.getByRole("button", { name: "Suggest 15 words" }).click();
  await expect(page.getByText("1 already in your deck, skipped")).toBeVisible();
  await expect(page.getByText("diagnosis", { exact: true })).toHaveCount(0);
  await expect(page.getByText("surgery", { exact: true })).toHaveCount(0);
});

test("a typed topic works", async ({ page, context, baseURL }) => {
  await signIn(context, baseURL!);
  await page.goto("/en/add-words?mode=topic");
  await page.getByRole("button", { name: "Other…" }).click();
  await page.getByLabel("Other…").fill("banking apps");
  await page.getByRole("button", { name: "Suggest 15 words" }).click();
  await expect(page.getByRole("heading", { name: "Words for “banking apps”" })).toBeVisible();
  await expect(page.getByText("15 words", { exact: true })).toBeVisible();
});
