type TUnitTranslator = (key: "lessThanMinute" | "minutes" | "hours" | "days" | "months" | "years", values?: { n: string }) => string;

const MINUTE = 60;
const HOUR = 60 * MINUTE;
const DAY = 24 * HOUR;

// Anki-style short interval: <1m, 6m, 3h, 4d, 2mo, 1.5y.
export function formatInterval(seconds: number, t: TUnitTranslator): string {
  if (seconds < MINUTE) {
    return t("lessThanMinute");
  }
  if (seconds < HOUR) {
    return t("minutes", { n: String(Math.round(seconds / MINUTE)) });
  }
  if (seconds < DAY) {
    return t("hours", { n: String(Math.round(seconds / HOUR)) });
  }
  const days = seconds / DAY;
  if (days < 30) {
    return t("days", { n: String(Math.round(days)) });
  }
  if (days < 365) {
    return t("months", { n: String(Math.round((days / 30) * 10) / 10) });
  }
  return t("years", { n: String(Math.round((days / 365) * 10) / 10) });
}
