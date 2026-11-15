"use client";

import { BookMarked } from "lucide-react";
import { useTranslations } from "next-intl";
import { cn } from "@/lib/utils";
import type { TFigure, TTask, TTaskSummary } from "../types";

export function SourceLabel({ source, className }: { source: TTaskSummary["source"]; className?: string }) {
  const t = useTranslations("writing");
  if (!source) return null;
  return (
    <p className={cn("inline-flex items-center gap-1.5 text-xs text-muted-foreground", className)}>
      <BookMarked className="size-3.5" strokeWidth={1.75} aria-hidden />
      {t("inspiredBy", { book: source.book, test: source.test })}
    </p>
  );
}

/** Figures are drawn on white, so they stay readable in dark mode. */
export function Figure({ figure, className }: { figure: TFigure; className?: string }) {
  return (
    // eslint-disable-next-line @next/next/no-img-element -- SVG served through the API's redirect to S3
    <img
      src={figure.url}
      alt={figure.alt}
      width={figure.width || undefined}
      height={figure.height || undefined}
      className={cn("h-auto w-full rounded-xl border border-border bg-white", className)}
      loading="lazy"
    />
  );
}

/** The exam question as the candidate sees it: time, the task box, the standard lines and the figure. */
export function TaskPrompt({ task, compact = false }: { task: TTask; compact?: boolean }) {
  const t = useTranslations("writing");
  const paragraphs = task.prompt.split(/\n{2,}/);
  return (
    <div className="space-y-4">
      {!compact && (
        <p className="text-sm text-muted-foreground">{t("spendAbout", { minutes: task.minutes })}</p>
      )}
      <div className="space-y-3 rounded-xl border-2 border-foreground/70 p-4 text-[15px] leading-relaxed italic">
        {task.task_no === 2 && <p className="not-italic text-muted-foreground">{t("writeAboutTopic")}</p>}
        {paragraphs.map((p, i) => (
          <p key={i} className="font-medium">
            {p}
          </p>
        ))}
        {task.task_no === 1 ? <p className="font-medium">{t("summariseLine")}</p> : <p>{t("giveReasonsLine")}</p>}
      </div>
      <p className="text-sm">{t("writeAtLeast", { words: task.min_words })}</p>
      {task.figures.map((f) => (
        <Figure key={f.url} figure={f} />
      ))}
    </div>
  );
}
