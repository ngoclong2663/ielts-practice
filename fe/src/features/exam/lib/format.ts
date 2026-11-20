/** 75 → "1:15"; 3725 → "1:02:05". */
export function formatClock(totalSec: number): string {
  const s = Math.max(0, Math.round(totalSec));
  const h = Math.floor(s / 3600);
  const m = Math.floor((s % 3600) / 60);
  const sec = String(s % 60).padStart(2, "0");
  return h ? `${h}:${String(m).padStart(2, "0")}:${sec}` : `${m}:${sec}`;
}

/** Suggested minutes: 20 per reading passage; listening follows its audio. */
export const READING_MINUTES_PER_PART = 20;
export const READING_FULL_MINUTES = 60;

export function partsParam(parts: number[]): string {
  return [...parts].sort((a, b) => a - b).join(",");
}

export function parseParts(raw: string | null): number[] {
  return (raw ?? "")
    .split(",")
    .map(Number)
    .filter((n) => Number.isInteger(n) && n >= 1 && n <= 4);
}
