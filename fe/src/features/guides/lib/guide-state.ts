import { useSyncExternalStore } from "react";

// Per-viewer memory in localStorage: the guides this browser has watched or put aside. A seen guide shrinks
// to a one-line "watch again" link. Losing it (private mode, cleared storage) only shows the big card again.
const KEY = "guides.seen";
const EVENT = "guides-seen";

function read(): string {
  try {
    return localStorage.getItem(KEY) ?? "";
  } catch {
    return "";
  }
}

function subscribe(onChange: () => void) {
  window.addEventListener(EVENT, onChange);
  window.addEventListener("storage", onChange);
  return () => {
    window.removeEventListener(EVENT, onChange);
    window.removeEventListener("storage", onChange);
  };
}

export function useGuideSeen(slug: string): boolean {
  const raw = useSyncExternalStore(subscribe, read, () => "");
  return raw.split(",").includes(slug);
}

export function markGuideSeen(slug: string) {
  const seen = new Set(read().split(",").filter(Boolean));
  if (seen.has(slug)) return;
  seen.add(slug);
  try {
    localStorage.setItem(KEY, [...seen].join(","));
  } catch {
    // blocked storage: the card simply stays
  }
  window.dispatchEvent(new Event(EVENT));
}
