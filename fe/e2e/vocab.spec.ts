import { expect, test } from "@playwright/test";
import { signIn } from "./helpers";

// Epic 3: paste words → generated cards → review → save → study.
// The e2e API runs with VOCAB_FAKE_PROVIDERS=true: deterministic content, inline SVG pictures,
// and entries starting with "zz" are treated as typos.

test.beforeEach(({ page }) => {
  page.on("pageerror", (error) => console.log(`[pageerror] ${error.message}\n${error.stack ?? ""}`));
});

const SHOTS = process.env.E2E_SCREENSHOTS;

test("paste words, review the generated cards, save and study them", async ({ page, context, baseURL }, info) => {
  await signIn(context, baseURL!);
  const shot = async (name: string) => {
    if (SHOTS) await page.screenshot({ path: `${SHOTS}/${info.project.name}-${name}.png`, fullPage: true });
  };

  await page.goto("/en/add-words");
  await page.getByLabel("Words or phrases").fill("decide, take into account\nzzbuzz\nDecide");
  // Duplicates are merged in the live preview.
  await expect(page.getByText("3 / 30")).toBeVisible();
  await expect(page.getByText("(duplicate)")).toBeVisible();
  await shot("20-paste");
  await page.getByRole("button", { name: "Generate 3 cards" }).click();

  await expect(page).toHaveURL(/job=/);
  await expect(page.getByRole("heading", { name: "Review your cards" })).toBeVisible();
  // The typo gets a suggestion; accepting it regenerates the word.
  await page.getByRole("button", { name: "Use “buzz”" }).click();
  await expect(page.getByText("3 of 3 cards ready")).toBeVisible({ timeout: 15_000 });
  await shot("21-review");

  // Preview: the same components as the study screen; the translation is hidden until tapped.
  const decideCard = page.getByRole("listitem").filter({ has: page.getByText("decide", { exact: true }) });
  await decideCard.getByRole("button", { name: "Preview & edit" }).last().click();
  const dialog = page.getByRole("dialog");
  await expect(dialog.getByText("Students often discuss decide in class.")).toBeVisible();
  await dialog.getByRole("button", { name: "Show translation" }).first().click();
  await expect(dialog.getByText("Học sinh thường thảo luận về decide trên lớp.")).toBeVisible();

  // Edit the user's copy.
  await dialog.getByRole("tab", { name: "Edit" }).click();
  await dialog.getByLabel("Vietnamese meaning").fill("quyết định");
  await dialog.getByRole("button", { name: "Save changes" }).click();
  await expect(page.getByText("Changes saved")).toBeVisible();
  await page.keyboard.press("Escape");
  await expect(decideCard.getByText("quyết định")).toBeVisible();

  // Change the picture.
  await page.getByRole("button", { name: "Change the picture for decide" }).click();
  const picker = page.getByRole("dialog");
  await expect(picker.getByRole("button", { name: "decide photo" }).first()).toBeVisible();
  await picker.getByRole("button", { name: "decide photo" }).nth(2).click();
  await expect(picker).toBeHidden();

  // Save into a new deck, then study.
  await page.getByRole("button", { name: "Save 3 cards" }).click();
  await expect(page.getByText(/Saved to Vocabulary/)).toBeVisible();
  await shot("22-saved");
  await page.getByRole("link", { name: "Start studying now" }).click();
  await expect(page.getByRole("button", { name: "Show answer" })).toBeVisible();

  // Generated cards can be reported from the study screen.
  await page.getByRole("button", { name: "Card options" }).click();
  await page.getByRole("menuitem", { name: "Report wrong content" }).click();
  await page.getByRole("radio", { name: "Bad example sentence" }).click();
  await page.getByRole("button", { name: "Send report" }).click();
  await expect(page.getByText("Thanks, we'll take a look.")).toBeVisible();
});

test("recent generations survive leaving the page", async ({ page, context, baseURL }) => {
  await signIn(context, baseURL!);
  await page.goto("/en/add-words");
  await page.getByLabel("Words or phrases").fill("sustainable");
  await page.getByRole("button", { name: "Generate 1 card" }).click();
  await expect(page.getByText("1 of 1 cards ready")).toBeVisible({ timeout: 15_000 });

  await page.getByRole("button", { name: "New list" }).click();
  await expect(page.getByRole("heading", { name: "Recent generations" })).toBeVisible();
  await page.getByRole("button", { name: /sustainable/ }).click();
  await expect(page.getByRole("heading", { name: "Review your cards" })).toBeVisible();
});

test("choose an existing deck while pasting", async ({ page, context, baseURL }) => {
  await signIn(context, baseURL!);
  await page.goto("/en/add-words");

  // Create a deck right from the paste box; it becomes the selected deck.
  await page.getByRole("button", { name: "or create a new deck" }).click();
  await page.getByLabel("Name").fill("Environment words");
  await page.getByRole("button", { name: "Create deck" }).click();
  await expect(page.getByRole("combobox", { name: "Save to" })).toHaveText(/Environment words/);

  await page.getByLabel("Words or phrases").fill("pollution, renewable energy");
  await page.getByRole("button", { name: "Generate 2 cards" }).click();
  await expect(page.getByText("2 of 2 cards ready")).toBeVisible({ timeout: 15_000 });

  // The save bar starts from the deck chosen while pasting.
  await expect(page.getByRole("combobox", { name: "Save to" })).toHaveText(/Environment words/);
  await page.getByRole("button", { name: "Save 2 cards" }).click();
  await expect(page.getByText("Saved to Environment words")).toBeVisible();
});
