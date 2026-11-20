import type { THighlight } from "../types";

// Highlights are painted with the CSS Custom Highlight API: ranges are registered on CSS.highlights and
// styled with ::highlight(), so the DOM React renders is never changed. Each highlight is stored as
// character offsets inside one text block marked with data-hl="<block id>".

export const HL_USER = "exam-user";
export const HL_EVIDENCE = "exam-evidence";

export function highlightsSupported(): boolean {
  return typeof window !== "undefined" && typeof CSS !== "undefined" && "highlights" in CSS && typeof Highlight !== "undefined";
}

/** Offset of a DOM position within a block's text. */
export function textOffset(block: Node, node: Node, offset: number): number {
  const r = document.createRange();
  r.selectNodeContents(block);
  r.setEnd(node, offset);
  return r.toString().length;
}

/** A DOM range for characters start..end of a block's text, or null if the offsets no longer fit. */
export function rangeFor(block: Element, start: number, end: number): Range | null {
  const walker = document.createTreeWalker(block, NodeFilter.SHOW_TEXT);
  const r = document.createRange();
  let pos = 0;
  let started = false;
  for (let node = walker.nextNode(); node; node = walker.nextNode()) {
    const len = node.nodeValue?.length ?? 0;
    if (!started && start <= pos + len) {
      r.setStart(node, start - pos);
      started = true;
    }
    if (started && end <= pos + len) {
      r.setEnd(node, end - pos);
      return r;
    }
    pos += len;
  }
  return null;
}

/** The current selection split into one piece per block it touches, inside root. */
export function selectionPieces(root: Element): { pieces: THighlight[]; rect: DOMRect } | null {
  const sel = window.getSelection();
  if (!sel || sel.rangeCount === 0 || sel.isCollapsed) return null;
  const r = sel.getRangeAt(0);
  const pieces: THighlight[] = [];
  root.querySelectorAll<HTMLElement>("[data-hl]").forEach((b) => {
    if (!r.intersectsNode(b)) return;
    const br = document.createRange();
    br.selectNodeContents(b);
    const start = r.compareBoundaryPoints(Range.START_TO_START, br) > 0 ? textOffset(b, r.startContainer, r.startOffset) : 0;
    const end = r.compareBoundaryPoints(Range.END_TO_END, br) < 0 ? textOffset(b, r.endContainer, r.endOffset) : (b.textContent?.length ?? 0);
    if (end > start && b.dataset.hl) pieces.push({ block: b.dataset.hl, start, end });
  });
  return pieces.length ? { pieces, rect: r.getBoundingClientRect() } : null;
}

/** Adds a highlight, merging it with any it overlaps in the same block. */
export function addHighlight(list: THighlight[], h: THighlight): THighlight[] {
  let { start, end } = h;
  const keep: THighlight[] = [];
  for (const x of list) {
    if (x.block !== h.block || x.end < start || x.start > end) keep.push(x);
    else {
      start = Math.min(start, x.start);
      end = Math.max(end, x.end);
    }
  }
  return [...keep, { block: h.block, start, end }];
}

/** The highlight at a caret position, if any. */
export function highlightAt(list: THighlight[], block: string, pos: number): THighlight | undefined {
  return list.find((h) => h.block === block && pos >= h.start && pos <= h.end);
}

/** Registers ranges under a highlight name (an empty list clears it). */
export function paint(name: string, root: Element | null, list: THighlight[]) {
  if (!highlightsSupported()) return;
  if (!root || list.length === 0) {
    CSS.highlights.delete(name);
    return;
  }
  const ranges: Range[] = [];
  for (const h of list) {
    const block = root.querySelector(`[data-hl="${CSS.escape(h.block)}"]`);
    const r = block && rangeFor(block, h.start, h.end);
    if (r) ranges.push(r);
  }
  CSS.highlights.set(name, new Highlight(...ranges));
}
