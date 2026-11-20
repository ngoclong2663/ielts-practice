import { expect, type Page, test } from "@playwright/test";
import { signIn } from "./helpers";

// Epic 10 Phase 1: pick one passage → answer (a gap, a word-limit warning, TRUE/FALSE/NOT GIVEN) →
// highlight text (kept in the browser only) → reload keeps everything → submit → score → review with the
// evidence painted in the passage.

test.beforeEach(({ page }) => {
  page.on("pageerror", (error) => console.log(`[pageerror] ${error.message}\n${error.stack ?? ""}`));
});

const SHOTS = process.env.E2E_SCREENSHOTS;

/** Selects the first `length` characters of a passage paragraph, as a mouse drag would. */
async function selectText(page: Page, block: string, length: number) {
  await page.evaluate(
    ([b, n]) => {
      const el = document.querySelector(`[data-hl="${b}"]`)!;
      const text = el.firstChild!;
      const r = document.createRange();
      r.setStart(text, 0);
      r.setEnd(text, Number(n));
      const sel = window.getSelection()!;
      sel.removeAllRanges();
      sel.addRange(r);
      document.dispatchEvent(new MouseEvent("mouseup", { bubbles: true }));
    },
    [block, String(length)],
  );
}

const highlightCount = (page: Page, name: string) =>
  page.evaluate((n) => (CSS.highlights.get(n) as Highlight | undefined)?.size ?? 0, name);

test("one reading passage from the test bank to the explanations", async ({ page, context, baseURL }, info) => {
  await signIn(context, baseURL!);
  const mobile = info.project.name === "mobile";
  const shot = async (name: string) => {
    if (SHOTS) await page.screenshot({ path: `${SHOTS}/${info.project.name}-${name}.png`, fullPage: true });
  };
  // Highlights must never reach the API (Epic 10 §5.2).
  const apiBodies: string[] = [];
  page.on("request", (req) => {
    if (req.url().includes("/api/v1/") && req.postData()) apiBodies.push(req.postData()!);
  });

  // The Reading section is in the navigation.
  await page.goto("/en/decks");
  const nav = page.getByRole("navigation", { name: "Main navigation" }).filter({ visible: true });
  if (mobile) {
    await nav.getByRole("link", { name: "Switch to Writing" }).click();
    await page.getByRole("navigation", { name: "Main navigation" }).filter({ visible: true }).getByRole("link", { name: "Switch to Reading" }).click();
  } else {
    await page.locator("#nav-section-reading").getByRole("link", { name: "Practice tests" }).click();
  }
  await expect(page).toHaveURL(/\/en\/reading$/);
  await expect(page.getByRole("heading", { level: 1, name: "Reading" })).toBeVisible();
  await expect(page.getByText("Marianne North: painter of the world's plants")).toBeVisible();
  await shot("60-reading-bank");

  // Tick Passage 1 of Test 1 and start with the suggested 20 minutes.
  await page.getByRole("checkbox", { name: /Marianne North/ }).click();
  const startBar = page.getByRole("region", { name: "Start practice" });
  await expect(startBar.getByText("1 part · 13 questions")).toBeVisible();
  await startBar.getByRole("link", { name: "Start" }).click();
  await expect(page).toHaveURL(/\/reading\/practice\/s1-t1\?parts=1&time=1200/);
  await expect(page.getByRole("heading", { level: 1, name: "Marianne North: painter of the world's plants" })).toBeVisible();
  await expect(page.getByRole("button", { name: /Time left/ })).toContainText(/(19|20):\d\d/);

  if (!mobile) {
    // Highlight the start of paragraph 1, then check it is painted and stored locally.
    await selectText(page, "p-p1", 30);
    await page.getByRole("toolbar", { name: "Highlight" }).getByRole("button", { name: "Highlight" }).click();
    await expect.poll(() => highlightCount(page, "exam-user")).toBe(1);
  }

  if (mobile) await page.getByRole("tab", { name: /Questions/ }).click();
  await page.getByRole("textbox", { name: "Question 1", exact: true }).fill("London");
  await page.getByRole("textbox", { name: "Question 6", exact: true }).fill("Charles Darwin");
  await expect(page.getByText("2 words")).toBeVisible();
  await page.getByRole("radiogroup", { name: "8" }).getByRole("radio", { name: "TRUE" }).click();
  await page.getByRole("button", { name: "Flag question 9 for review" }).click();
  await shot("61-reading-player");

  // Everything survives a reload (the run lives in localStorage).
  await page.reload();
  if (mobile) await page.getByRole("tab", { name: /Questions/ }).click();
  await expect(page.getByRole("textbox", { name: "Question 1", exact: true })).toHaveValue("London");
  await expect(page.getByRole("radiogroup", { name: "8" }).getByRole("radio", { name: "TRUE" })).toHaveAttribute("aria-checked", "true");
  if (!mobile) await expect.poll(() => highlightCount(page, "exam-user")).toBe(1);

  await page.getByRole("button", { name: "Submit" }).first().click();
  const dialog = page.getByRole("alertdialog");
  await expect(dialog.getByText("10 questions are still unanswered.")).toBeVisible();
  await expect(dialog.getByText("1 question is flagged for review.")).toBeVisible();
  await dialog.getByRole("button", { name: "Submit" }).click();

  // Result: London and TRUE are right; "Charles Darwin" breaks ONE WORD ONLY.
  await expect(page).toHaveURL(/\/reading\/attempts\/[0-9a-f-]{36}$/);
  await expect(page.getByRole("heading", { level: 1, name: "Result · Passage 1" })).toBeVisible();
  await expect(page.getByText("/13 correct")).toBeVisible();
  await expect(page.getByText("A band is given only for a full test of 40 questions.")).toBeVisible();
  await expect(page.getByText("over the word limit")).toBeVisible();
  await expect(page.getByRole("link", { name: "Redo 11 wrong answers" })).toBeVisible();
  await shot("62-reading-result");

  // Review: the explanation for question 1 and its evidence in the passage.
  await page.getByRole("link", { name: "Review with explanations" }).click();
  await expect(page).toHaveURL(/\/review$/);
  await expect(page.getByText("The notes list Hastings and Norfolk")).toBeVisible();
  await page.getByRole("button", { name: "Show in passage · paragraph 1" }).first().click();
  await expect.poll(() => highlightCount(page, "exam-evidence")).toBe(1);
  if (!mobile) await expect.poll(() => highlightCount(page, "exam-user")).toBe(1);
  await shot("63-reading-review");

  expect(apiBodies.some((b) => b.includes("highlight") || b.includes("p-p1"))).toBe(false);
});
