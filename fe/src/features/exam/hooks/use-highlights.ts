import { type RefObject, useCallback, useEffect, useState } from "react";
import {
  HL_USER,
  addHighlight,
  highlightAt,
  highlightsSupported,
  paint,
  rangeFor,
  selectionPieces,
  textOffset,
} from "../lib/highlights";
import type { THighlight } from "../types";

export type TToolbar =
  | { kind: "add"; x: number; y: number; below: boolean; pieces: THighlight[] }
  | { kind: "remove"; x: number; y: number; below: boolean; target: THighlight }
  | null;

const isCoarse = () => typeof window !== "undefined" && window.matchMedia("(pointer: coarse)").matches;

function place(rect: DOMRect): { x: number; y: number; below: boolean } {
  // On touch screens the system menu sits above the selection, so the toolbar goes below it.
  const below = isCoarse() || rect.top < 64;
  return { x: rect.left + rect.width / 2, y: below ? rect.bottom + 10 : rect.top - 10, below };
}

/**
 * Highlighting kept in the browser (Epic 10 §5.2). Select text inside root (blocks marked data-hl) to get
 * an "add" toolbar; click inside a highlight to get a "remove" one; ⌘/Ctrl+Shift+H highlights the selection.
 * `repaint` is any value that changes when the text layout changes (pane switch, font size).
 */
export function useHighlights(
  root: RefObject<HTMLElement | null>,
  highlights: THighlight[],
  onChange: (next: THighlight[]) => void,
  repaint: unknown = null,
) {
  const [toolbar, setToolbar] = useState<TToolbar>(null);
  const supported = highlightsSupported();

  useEffect(() => {
    paint(HL_USER, root.current, highlights);
  }, [root, highlights, repaint]);

  useEffect(() => () => paint(HL_USER, null, []), []);

  const add = useCallback(
    (pieces: THighlight[]) => {
      onChange(pieces.reduce(addHighlight, highlights));
      window.getSelection()?.removeAllRanges();
      setToolbar(null);
    },
    [highlights, onChange],
  );

  const removeOne = useCallback(
    (target: THighlight) => {
      onChange(highlights.filter((h) => h !== target));
      setToolbar(null);
    },
    [highlights, onChange],
  );

  useEffect(() => {
    if (!supported) return;
    const check = () => {
      const el = root.current;
      if (!el) return;
      const sel = selectionPieces(el);
      if (sel) {
        setToolbar({ kind: "add", ...place(sel.rect), pieces: sel.pieces });
        return;
      }
      // A plain click inside an existing highlight offers to remove it.
      const s = window.getSelection();
      if (s && s.rangeCount && s.isCollapsed) {
        const r = s.getRangeAt(0);
        const node = r.startContainer.nodeType === Node.TEXT_NODE ? r.startContainer.parentElement : (r.startContainer as Element);
        const block = node?.closest<HTMLElement>("[data-hl]");
        if (block && el.contains(block) && block.dataset.hl) {
          const hit = highlightAt(highlights, block.dataset.hl, textOffset(block, r.startContainer, r.startOffset));
          const hr = hit && rangeFor(block, hit.start, hit.end);
          if (hit && hr) {
            setToolbar({ kind: "remove", ...place(hr.getBoundingClientRect()), target: hit });
            return;
          }
        }
      }
      setToolbar(null);
    };
    const onUp = () => window.setTimeout(check, 0);
    const onKeyUp = (e: KeyboardEvent) => {
      if (e.shiftKey || e.key === "Escape") window.setTimeout(check, 0);
    };
    const onKeyDown = (e: KeyboardEvent) => {
      if ((e.metaKey || e.ctrlKey) && e.shiftKey && e.key.toLowerCase() === "h") {
        const el = root.current;
        const sel = el && selectionPieces(el);
        if (sel) {
          e.preventDefault();
          add(sel.pieces);
        }
      }
    };
    let timer = 0;
    const onSelChange = () => {
      if (!isCoarse()) return;
      window.clearTimeout(timer);
      timer = window.setTimeout(check, 350);
    };
    const hide = () => setToolbar(null);
    document.addEventListener("mouseup", onUp);
    document.addEventListener("keyup", onKeyUp);
    document.addEventListener("keydown", onKeyDown);
    document.addEventListener("selectionchange", onSelChange);
    document.addEventListener("scroll", hide, true);
    return () => {
      window.clearTimeout(timer);
      document.removeEventListener("mouseup", onUp);
      document.removeEventListener("keyup", onKeyUp);
      document.removeEventListener("keydown", onKeyDown);
      document.removeEventListener("selectionchange", onSelChange);
      document.removeEventListener("scroll", hide, true);
    };
  }, [supported, root, highlights, add]);

  return { supported, toolbar, add, removeOne, dismiss: () => setToolbar(null) };
}
