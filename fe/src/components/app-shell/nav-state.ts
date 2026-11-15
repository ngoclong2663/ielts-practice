import { useSyncExternalStore } from "react";
import type { TSectionId } from "./nav-items";

// Per-viewer navigation preferences in localStorage: collapsed sidebar sections and the last
// section visited (the mobile bar shows it on pages outside any section, such as the dashboard).
const COLLAPSED_KEY = "nav.collapsed";
const LAST_SECTION_KEY = "nav.lastSection";
const EVENT = "nav-state";

function read(key: string): string {
  try {
    return localStorage.getItem(key) ?? "";
  } catch {
    return "";
  }
}

function write(key: string, value: string) {
  try {
    localStorage.setItem(key, value);
  } catch {
    // private mode or blocked storage: keep the default
  }
  window.dispatchEvent(new Event(EVENT));
}

function subscribe(onChange: () => void) {
  window.addEventListener(EVENT, onChange);
  window.addEventListener("storage", onChange);
  return () => {
    window.removeEventListener(EVENT, onChange);
    window.removeEventListener("storage", onChange);
  };
}

/** Comma-separated ids of collapsed sections ("" on the server: everything open). */
export function useCollapsedSections(): Set<string> {
  const raw = useSyncExternalStore(subscribe, () => read(COLLAPSED_KEY), () => "");
  return new Set(raw.split(",").filter(Boolean));
}

export function setSectionCollapsed(id: TSectionId, collapsed: boolean) {
  const set = new Set(read(COLLAPSED_KEY).split(",").filter(Boolean));
  if (collapsed) set.add(id);
  else set.delete(id);
  write(COLLAPSED_KEY, [...set].join(","));
}

export function useLastSection(): TSectionId {
  const raw = useSyncExternalStore(subscribe, () => read(LAST_SECTION_KEY), () => "");
  return raw === "writing" ? "writing" : "vocabulary";
}

export function rememberSection(id: TSectionId) {
  if (read(LAST_SECTION_KEY) !== id) write(LAST_SECTION_KEY, id);
}
