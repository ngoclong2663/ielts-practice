import { useState } from "react";

// Unsent text is kept in localStorage per attempt, so a reload or a failed AI call never loses it.
// It is a convenience only: the server keeps everything that was checked or submitted.
const PREFIX = "writing.draft.";

function read<T>(key: string): T | undefined {
  try {
    const raw = localStorage.getItem(PREFIX + key);
    return raw ? (JSON.parse(raw) as T) : undefined;
  } catch {
    return undefined;
  }
}

export function clearDraft(key: string) {
  try {
    localStorage.removeItem(PREFIX + key);
  } catch {
    // storage blocked: nothing to clear
  }
}

/** useState that also saves to localStorage; the stored draft wins over `initial` on first render. */
export function useDraft<T>(key: string, initial: T): [T, (value: T) => void] {
  const [value, setValue] = useState<T>(() => (typeof window === "undefined" ? initial : (read<T>(key) ?? initial)));
  const set = (next: T) => {
    setValue(next);
    try {
      localStorage.setItem(PREFIX + key, JSON.stringify(next));
    } catch {
      // storage blocked: the draft lives in memory only
    }
  };
  return [value, set];
}
