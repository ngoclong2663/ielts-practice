import path from "node:path";
import { expect, type Page, test } from "@playwright/test";
import { signIn } from "./helpers";

// Epic 10 Phase 2: pick one listening part → practice mode (play, answer) → submit → score → review with the
// transcript and "listen again"; exam mode plays once and stops for the reading time.
// S3 is off in e2e, so every recording is served from a 20-second silent fixture.

const SHOTS = process.env.E2E_SCREENSHOTS;
const SILENCE = path.join(__dirname, "fixtures", "silence-20s.mp3");

test.beforeEach(async ({ page }) => {
  page.on("pageerror", (error) => console.log(`[pageerror] ${error.message}\n${error.stack ?? ""}`));
  await page.route("**/api/v1/exam/media/**", (route) => route.fulfill({ path: SILENCE, contentType: "audio/mpeg" }));
});

async function shot(page: Page, project: string, name: string) {
  if (SHOTS) await page.screenshot({ path: `${SHOTS}/${project}-${name}.png`, fullPage: true });
}

test("one listening part in practice mode, then the transcript review", async ({ page, context, baseURL }, info) => {
  await signIn(context, baseURL!);
  const mobile = info.project.name === "mobile";

  await page.goto("/en/decks");
  if (!mobile) {
    await page.locator("#nav-section-listening").getByRole("link", { name: "Practice tests" }).click();
  } else {
    await page.goto("/en/listening");
  }
  await expect(page).toHaveURL(/\/en\/listening$/);
  await expect(page.getByRole("heading", { level: 1, name: "Listening" })).toBeVisible();
  await page.getByRole("checkbox", { name: /Courses at a climbing centre/ }).click();
  const startBar = page.getByRole("region", { name: "Start practice" });
  await expect(startBar.getByText("1 part · 10 questions")).toBeVisible();
  await shot(page, info.project.name, "70-listening-bank");
  await startBar.getByRole("link", { name: "Start" }).click();
  await expect(page).toHaveURL(/\/listening\/practice\/s1-t1\?parts=1&mode=practice/);

  // Practice mode: play, pause and speed are all there; the transcript is not shown yet.
  await page.getByRole("button", { name: "Play" }).click();
  await expect(page.getByRole("button", { name: "Pause" })).toBeVisible();
  await expect(page.getByRole("radio", { name: "1.25×" })).toBeVisible();
  await expect(page.getByText("Good afternoon, Lakeside Climbing Centre")).toHaveCount(0);

  await page.getByRole("textbox", { name: "Question 1", exact: true }).fill("harness");
  await page.getByRole("textbox", { name: "Question 2", exact: true }).fill("8");
  await page.getByRole("textbox", { name: "Question 6", exact: true }).fill("Tuesday");
  await shot(page, info.project.name, "71-listening-player");

  await page.getByRole("button", { name: "Submit" }).first().click();
  const dialog = page.getByRole("alertdialog");
  await expect(dialog.getByText("The recording has not finished yet.")).toBeVisible();
  await dialog.getByRole("button", { name: "Submit" }).click();

  await expect(page).toHaveURL(/\/listening\/attempts\/[0-9a-f-]{36}$/);
  await expect(page.getByRole("heading", { level: 1, name: "Result · Part 1" })).toBeVisible();
  await expect(page.getByText("/10 correct")).toBeVisible();
  await shot(page, info.project.name, "72-listening-result");

  // Review: the transcript with the answer words, and "listen again" plays from the evidence.
  await page.getByRole("link", { name: "Review with explanations" }).click();
  await expect(page).toHaveURL(/\/review$/);
  await expect(page.getByText("Tuesday was the old day.")).toBeVisible();
  await page.getByRole("button", { name: /Listen again/ }).first().click();
  await expect(page.getByRole("button", { name: "Pause" })).toBeVisible();
  await expect(page.getByText("Good afternoon, Lakeside Climbing Centre, Sophie speaking.")).toBeVisible();
  await shot(page, info.project.name, "73-listening-review");
});

test("exam mode plays once and stops for the reading time", async ({ page, context, baseURL }, info) => {
  test.skip(info.project.name === "mobile", "covered on desktop");
  await signIn(context, baseURL!);
  await page.goto("/en/listening/practice/s1-t1?parts=1&mode=exam");
  await expect(page.getByText("Plays once. You cannot pause or rewind.")).toBeVisible();
  await expect(page.getByRole("radio", { name: "1.25×" })).toHaveCount(0);
  await page.getByRole("button", { name: "Start the recording" }).click();
  // The first reading pause comes after the narrator's introduction (about 10 seconds in).
  await expect(page.getByRole("status").filter({ hasText: "Reading time: Questions 1–6" })).toBeVisible({ timeout: 25_000 });
  await expect(page.getByRole("button", { name: "Skip" })).toHaveCount(0);
  await shot(page, info.project.name, "74-listening-exam-cue");
});
