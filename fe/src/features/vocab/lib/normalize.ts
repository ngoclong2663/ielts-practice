import type { TEntry, TRejectedEntry } from "../types";

// Client copy of the backend rules (be/app/modules/vocab/domain/normalize.go) for live chips.
// The server normalises again when the job is created.
const MAX_CHARS = 60;
const MAX_WORDS = 6;
const MAX_HINT = 40;

const LIST_MARKER = /^\s*(?:[-*•·]+|\d+[.)])\s*/u;
const PIPE_HINT = /^(.+?)\s*\|\s*(.+)$/u;
const PAREN_HINT = /^(.+?)\s*\(([^()]+)\)$/u;
const ALLOWED = /^\p{L}[\p{L}\p{M}' .\-’]*$/u;

function lowerUnlessAcronym(word: string) {
  const upper = [...word].filter((c) => c !== c.toLowerCase() && c === c.toUpperCase()).length;
  return upper >= 2 ? word : word.toLowerCase();
}

export const lexemeKey = (text: string, hint: string) => `${text.trim().toLowerCase()}|${hint.trim().toLowerCase()}`;

export function normalizeEntries(text: string): { entries: TEntry[]; rejected: TRejectedEntry[] } {
  const entries: TEntry[] = [];
  const rejected: TRejectedEntry[] = [];
  const seen = new Set<string>();
  for (const part of text.split(/[\n\r,;]+/u)) {
    const raw = part.replace(LIST_MARKER, "").replace(/[’‘]/gu, "'").replace(/\s+/gu, " ").trim();
    if (!raw) continue;
    let word = raw;
    let hint = "";
    const m = PIPE_HINT.exec(raw) ?? PAREN_HINT.exec(raw);
    if (m) {
      word = m[1];
      hint = m[2];
    }
    word = word.trim().replace(/^\.+|\.+$/gu, "");
    hint = [...hint.trim().toLowerCase()].slice(0, MAX_HINT).join("");
    if ([...word].length > MAX_CHARS) {
      rejected.push({ text: raw, reason: "TOO_LONG" });
      continue;
    }
    if (word.split(" ").filter(Boolean).length > MAX_WORDS) {
      rejected.push({ text: raw, reason: "TOO_MANY_WORDS" });
      continue;
    }
    if (!ALLOWED.test(word)) {
      rejected.push({ text: raw, reason: "INVALID_CHARS" });
      continue;
    }
    word = lowerUnlessAcronym(word);
    const key = lexemeKey(word, hint);
    if (seen.has(key)) {
      rejected.push({ text: raw, reason: "DUPLICATE" });
      continue;
    }
    seen.add(key);
    entries.push({ text: word, sense_hint: hint });
  }
  return { entries, rejected };
}
