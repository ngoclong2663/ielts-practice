"use client";

import { useRef } from "react";
import type { KeyboardEvent, PointerEvent } from "react";

/**
 * Shift-click range selection for tick-box lists, as in mail apps: click one item, then Shift-click another, and
 * every item between them takes the new state of the one Shift-clicked. `keys` are the visible items in order.
 *
 * Spread `rowProps` on each row (it notes whether Shift was held and stops the text selection Shift-click makes),
 * then call `affected(key)` in the toggle handler to get the keys to change.
 */
export function useRangeSelect(keys: string[]) {
  const anchor = useRef<string | null>(null);
  const shift = useRef(false);

  const rowProps = {
    onPointerDownCapture: (e: PointerEvent) => {
      shift.current = e.shiftKey;
      if (e.shiftKey) e.preventDefault();
    },
    onKeyDownCapture: (e: KeyboardEvent) => {
      shift.current = e.shiftKey;
    },
  };

  const affected = (key: string): string[] => {
    const from = anchor.current;
    const range = shift.current;
    shift.current = false;
    anchor.current = key;
    const a = from ? keys.indexOf(from) : -1;
    const b = keys.indexOf(key);
    if (!range || a < 0 || b < 0) return [key];
    return keys.slice(Math.min(a, b), Math.max(a, b) + 1);
  };

  return { rowProps, affected };
}
