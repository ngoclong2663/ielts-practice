// The deck the user last saved generated cards into (a per-browser convenience only).
const LAST_DECK_KEY = "vocab:last-deck";

export function readLastDeck(): string | null {
  try {
    return window.localStorage.getItem(LAST_DECK_KEY);
  } catch {
    return null;
  }
}

export function rememberDeck(id: string) {
  try {
    window.localStorage.setItem(LAST_DECK_KEY, id);
  } catch {
    // Storage can be unavailable (private mode); the default deck is only a convenience.
  }
}
