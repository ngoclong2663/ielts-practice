import { expect, test } from "@playwright/test";
import { signIn } from "./helpers";

// Epic 5: pick a Task 1 → understand it (Step 0) → write the introduction sentence by sentence →
// write the full answer from those sentences → feedback with an estimated band.
// The e2e API uses VOCAB_FAKE_PROVIDERS=true: the fake coach writes a fixed lesson, treats
// "drink medicine" as a word-for-word translation and gives 6.0 per criterion to 150+ words.

test.beforeEach(({ page }) => {
  page.on("pageerror", (error) => console.log(`[pageerror] ${error.message}\n${error.stack ?? ""}`));
});

const SHOTS = process.env.E2E_SCREENSHOTS;

const filler = (n: number) =>
  Array.from({ length: n }, (_, i) => (i % 12 === 11 ? "steadily." : "workers")).join(" ");

test("guided Task 1 practice from the plan to the band estimate", async ({ page, context, baseURL }, info) => {
  await signIn(context, baseURL!);
  const shot = async (name: string) => {
    if (SHOTS) await page.screenshot({ path: `${SHOTS}/${info.project.name}-${name}.png`, fullPage: true });
  };
  const mobile = info.project.name === "mobile";

  // The navigation has a Writing section (sidebar on desktop, section switch on mobile).
  await page.goto("/en/decks");
  const nav = page.getByRole("navigation", { name: "Main navigation" }).filter({ visible: true });
  if (mobile) {
    await nav.getByRole("link", { name: "Switch to Writing" }).click();
  } else {
    await nav.getByRole("link", { name: "Task 1 (charts)" }).click();
  }
  await expect(page).toHaveURL(/\/writing\/task-1$/);
  await expect(page.getByRole("heading", { name: "Writing Task 1" })).toBeVisible();
  await expect(page.getByRole("link", { name: /Workers in four industries/ })).toBeVisible();
  await shot("50-writing-tasks");

  await page.getByRole("link", { name: /Workers in four industries/ }).click();
  await expect(page).toHaveURL(/\/writing\/tasks\//);
  await expect(page.getByRole("heading", { level: 1, name: "Workers in four industries, 1965–2025" })).toBeVisible();
  // Next keeps the previous page mounted but hidden, so match only what is visible.
  // The reference book is never named (source_ref is internal).
  await expect(page.getByText(/Cambridge/)).toHaveCount(0);
  await expect(page.getByText("Summarise the information by selecting and reporting the main features").first()).toBeVisible();
  await expect(page.getByRole("radio", { name: /Full support/ })).toBeChecked();
  await shot("51-writing-task");
  await page.getByRole("button", { name: "Start practising" }).click();

  // Step 0: understand the task.
  await expect(page).toHaveURL(/\/writing\/attempts\//);
  await expect(page.getByRole("heading", { name: "Understand the task" })).toBeVisible();
  await page.getByRole("radio", { name: "Mô tả và so sánh" }).check();
  await page.getByRole("radio", { name: "150" }).check();
  await page.getByRole("button", { name: "Check my answers" }).click();
  await expect(page.getByText("The plan of the answer")).toBeVisible();
  await page.getByRole("button", { name: "Continue" }).click();

  // Introduction: one sentence with a word-for-word translation, one left empty.
  await expect(page.getByRole("heading", { name: "Introduction" })).toBeVisible();
  await expect(page.getByText("Câu 1 của đoạn introduction.")).toBeVisible();
  await page.getByRole("button", { name: "Show hints" }).first().click();
  await expect(page.getByText("take medicine").first()).toBeVisible();
  await page.getByLabel("Sentence 1").fill("Old people drink medicine every day at home.");
  await page.getByRole("button", { name: "Check paragraph" }).click();
  await expect(page.getByText("Understandable, with mistakes")).toBeVisible();
  await expect(page.getByText("Word-for-word translation")).toBeVisible();
  await expect(page.getByText("Old people take medicine every day at home.")).toBeVisible();
  await expect(page.getByText("Not written")).toBeVisible();
  await expect(page.getByText("39 of 40 checks left today")).toBeVisible();
  await shot("52-writing-paragraph");

  // Jump to the full answer; start from the checked (corrected) sentences.
  await page.getByRole("button", { name: /Full answer/ }).click();
  await expect(page.getByText("Your sentences so far")).toBeVisible();
  await page.getByRole("button", { name: "Start from my sentences" }).click();
  const editor = page.getByRole("textbox", { name: "Write the full answer" });
  await expect(editor).toHaveValue("Old people take medicine every day at home.");
  await editor.fill(`Old people take medicine every day at home.\n\n${filler(160)}`);
  await expect(page.getByText(/^16\d \/ 150 words$/)).toBeVisible();
  await shot("53-writing-final");
  await page.getByRole("button", { name: "Submit for feedback" }).click();

  await expect(page.getByTestId("band")).toHaveText("5.5–6.0");
  await expect(page.getByRole("heading", { name: "Task Achievement" })).toBeVisible();
  await expect(page.getByText("Three fixes that help most")).toBeVisible();
  await expect(page.getByText("Model answer")).toBeVisible();
  await shot("54-writing-result");

  // The attempt is in the history with its band.
  await page.goto("/en/writing/history");
  await expect(page.getByRole("link", { name: /Workers in four industries/ })).toContainText("Band 5.5–6.0");
});

test("Task 2 asks for a position before starting", async ({ page, context, baseURL }) => {
  await signIn(context, baseURL!, "vi");
  await page.goto("/vi/writing/task-2");
  await page.getByRole("link", { name: /High-rise flats/ }).click();
  await expect(page).toHaveURL(/\/writing\/tasks\//);
  const start = page.getByRole("button", { name: "Bắt đầu luyện" });
  await expect(start).toBeDisabled();
  await page.getByRole("button", { name: "Đồng ý một phần" }).click();
  await start.click();
  await expect(page.getByRole("heading", { name: "Hiểu đề bài" })).toBeVisible();
  await expect(page.getByText("Đồng ý một phần")).toBeVisible();
  // Task 2 has a conclusion step and no overview.
  await expect(page.getByRole("button", { name: /Kết bài/ })).toBeVisible();
  await expect(page.getByRole("button", { name: /Tổng quan/ })).toHaveCount(0);
});
