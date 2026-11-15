/** IELTS bands are shown with one decimal: 6 → "6.0". */
export const formatBand = (band: number) => band.toFixed(1);

export const formatBandRange = (low: number, high: number) =>
  low === high ? formatBand(high) : `${formatBand(low)}–${formatBand(high)}`;

/** 754 → "12:34". */
export function formatClock(totalSeconds: number) {
  const s = Math.max(0, Math.floor(totalSeconds));
  return `${Math.floor(s / 60)}:${String(s % 60).padStart(2, "0")}`;
}

export const countWords = (text: string) => text.trim().split(/\s+/).filter(Boolean).length;
