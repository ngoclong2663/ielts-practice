"use client";

import { type PointerEvent as ReactPointerEvent, type ReactNode, useRef, useState } from "react";
import { useTranslations } from "next-intl";
import { cn } from "@/lib/utils";

const WIDTH_KEY = "exam.split";
const MIN = 32;
const MAX = 72;

function readWidth(): number {
  try {
    const saved = Number(localStorage.getItem(WIDTH_KEY));
    return saved >= MIN && saved <= MAX ? saved : 55;
  } catch {
    return 55;
  }
}

type SplitViewProps = {
  left: ReactNode;
  right: ReactNode;
  /** On a phone only one pane shows; this picks it. */
  show: "left" | "right";
  leftLabel: string;
  rightLabel: string;
  onResize?: () => void;
};

/** Passage | questions, each scrolling on its own, resizable from 32 % to 72 % (Epic 10 §5.1). */
export function SplitView({ left, right, show, leftLabel, rightLabel, onResize }: SplitViewProps) {
  const t = useTranslations("exam");
  const box = useRef<HTMLDivElement>(null);
  // The player renders only in the browser (after its data loads), so storage can be read on first render.
  const [width, setWidth] = useState(readWidth);

  const commit = (w: number) => {
    const v = Math.min(MAX, Math.max(MIN, w));
    setWidth(v);
    try {
      localStorage.setItem(WIDTH_KEY, String(v));
    } catch {
      // storage blocked
    }
    onResize?.();
  };

  const onPointerDown = (e: ReactPointerEvent<HTMLDivElement>) => {
    const el = e.currentTarget;
    el.setPointerCapture(e.pointerId);
    const move = (ev: PointerEvent) => {
      const r = box.current?.getBoundingClientRect();
      if (r) setWidth(Math.min(MAX, Math.max(MIN, ((ev.clientX - r.left) / r.width) * 100)));
    };
    const up = (ev: PointerEvent) => {
      el.removeEventListener("pointermove", move);
      const r = box.current?.getBoundingClientRect();
      if (r) commit(((ev.clientX - r.left) / r.width) * 100);
    };
    el.addEventListener("pointermove", move);
    el.addEventListener("pointerup", up, { once: true });
  };

  return (
    <div ref={box} className="relative min-h-0 flex-1 md:grid" style={{ gridTemplateColumns: `${width}% 8px minmax(0,1fr)` }}>
      <section
        aria-label={leftLabel}
        className={cn("absolute inset-0 overflow-y-auto bg-card px-4 pt-5 pb-20 md:static md:px-9 md:pt-7", show === "right" && "hidden md:block")}
      >
        {left}
      </section>
      <div
        role="separator"
        aria-orientation="vertical"
        aria-label={t("resize")}
        aria-valuenow={Math.round(width)}
        aria-valuemin={MIN}
        aria-valuemax={MAX}
        tabIndex={0}
        onPointerDown={onPointerDown}
        onKeyDown={(e) => {
          if (e.key === "ArrowLeft") commit(width - 2);
          if (e.key === "ArrowRight") commit(width + 2);
        }}
        className="relative hidden cursor-col-resize bg-border outline-none after:absolute after:top-1/2 after:left-0.5 after:h-9 after:w-1 after:-translate-y-1/2 after:rounded-full after:bg-input focus-visible:bg-ring md:block"
      />
      <section
        aria-label={rightLabel}
        className={cn("absolute inset-0 overflow-y-auto px-4 pt-4 pb-20 md:static md:px-7 md:pt-6", show === "left" && "hidden md:block")}
      >
        {right}
      </section>
    </div>
  );
}
