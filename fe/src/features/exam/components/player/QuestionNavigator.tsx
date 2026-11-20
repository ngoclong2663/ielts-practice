"use client";

import { ChevronLeft, ChevronRight } from "lucide-react";
import { useTranslations } from "next-intl";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";
import { range } from "../../lib/questions";
import type { TPart, TSkill } from "../../types";

type QuestionNavigatorProps = {
  skill: TSkill;
  parts: TPart[];
  current: number;
  /** Which questions count in this run (all, or the redo set). */
  inRun: (n: number) => boolean;
  answered: (n: number) => boolean;
  flagged: (n: number) => boolean;
  /** Review: question → correct. */
  correct?: (n: number) => boolean | undefined;
  onGo: (n: number) => void;
  onStep: (delta: number) => void;
};

/** The bottom bar: a tab per part with its answered count, a chip per question, previous/next (Epic 10 §5.1). */
export function QuestionNavigator({ skill, parts, current, inRun, answered, flagged, correct, onGo, onStep }: QuestionNavigatorProps) {
  const t = useTranslations("exam");
  const all = parts.flatMap((p) => range(p.q_from, p.q_to)).filter(inRun);
  const pos = Math.max(0, all.indexOf(current));
  return (
    <nav
      aria-label={t("navigator")}
      className="flex min-h-14 shrink-0 items-center gap-2 border-t border-border bg-card px-3 py-2 pb-[max(0.5rem,env(safe-area-inset-bottom))]"
    >
      <div className="hidden min-w-0 flex-1 items-center gap-3 overflow-x-auto px-1 py-1.5 md:flex">
        {parts.map((p) => {
          const qs = range(p.q_from, p.q_to).filter(inRun);
          if (!qs.length) return null;
          const done = qs.filter(answered).length;
          return (
            <div key={p.part_no} className="flex shrink-0 items-center gap-1.5">
              <button
                type="button"
                onClick={() => onGo(qs[0])}
                className={cn(
                  "h-8 shrink-0 rounded-md px-2.5 text-[13px] font-semibold",
                  current >= p.q_from && current <= p.q_to ? "bg-muted" : "text-muted-foreground hover:bg-muted",
                )}
              >
                {t(`partLabel.${skill}`, { n: p.part_no })}
                {!correct && <span className="ml-1.5 font-normal text-muted-foreground tabular-nums">{done}/{qs.length}</span>}
              </button>
              <div className="flex gap-1">
                {qs.map((n) => {
                  const c = correct?.(n);
                  return (
                    <button
                      key={n}
                      type="button"
                      onClick={() => onGo(n)}
                      aria-label={t("goTo", { n })}
                      aria-current={current === n ? "true" : undefined}
                      className={cn(
                        "relative size-8 shrink-0 rounded-md border border-input bg-background text-[13px] font-medium tabular-nums",
                        c === undefined && answered(n) && "border-transparent bg-accent text-accent-foreground",
                        c === true && "border-transparent bg-accent text-accent-foreground",
                        c === false && "border-transparent bg-destructive/10 text-destructive",
                        current === n && "border-primary ring-1 ring-primary ring-inset",
                        flagged(n) &&
                          "after:absolute after:-top-1 after:-right-1 after:size-2.5 after:rounded-full after:border-2 after:border-card after:bg-amber-500",
                      )}
                    >
                      {n}
                    </button>
                  );
                })}
              </div>
            </div>
          );
        })}
      </div>
      <div className="flex flex-1 items-center justify-between gap-2 md:flex-none">
        <Button variant="outline" size="icon-lg" onClick={() => onStep(-1)} disabled={pos === 0} aria-label={t("previous")}>
          <ChevronLeft aria-hidden />
        </Button>
        <span className="text-sm tabular-nums md:hidden">{t("position", { n: current, pos: pos + 1, total: all.length })}</span>
        <Button variant="outline" size="icon-lg" onClick={() => onStep(1)} disabled={pos === all.length - 1} aria-label={t("next")}>
          <ChevronRight aria-hidden />
        </Button>
      </div>
    </nav>
  );
}
