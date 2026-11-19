"use client";

import { ArrowRight, LoaderCircle, RotateCcw } from "lucide-react";
import { useTranslations } from "next-intl";
import { Button } from "@/components/ui/button";
import { useRouter } from "@/i18n/navigation";
import { useStartAttempt } from "../hooks/use-writing";
import { toastApiError } from "../lib/errors";
import { formatBand, formatBandRange, formatClock } from "../lib/format";
import type { TAttempt, TLevel, TPosition } from "../types";
import { UsefulVocab } from "./UsefulVocab";

export function ResultView({ attempt }: { attempt: TAttempt }) {
  const t = useTranslations();
  const router = useRouter();
  const again = useStartAttempt(attempt.task.id);
  const s = attempt.submission;
  if (!s) return null;
  const fb = s.feedback;

  const restart = (level: TLevel) =>
    again.mutate(
      { level, ...(attempt.position ? { position: attempt.position as TPosition } : {}) },
      { onSuccess: (a) => router.push(`/writing/attempts/${a.id}`), onError: (e) => toastApiError(t, e, "writing-start") },
    );

  return (
    <section className="space-y-6">
      <div className="rounded-2xl border border-primary/30 bg-primary/5 p-5 text-center">
        <p className="text-sm text-muted-foreground">{t("writing.result.estimated")}</p>
        <p className="mt-1 text-4xl font-semibold tabular-nums" data-testid="band">
          {formatBandRange(s.band_low, s.band_high)}
        </p>
        <p className="mx-auto mt-2 max-w-md text-xs text-muted-foreground">{t("writing.result.disclaimer")}</p>
      </div>

      {fb.off_task && (
        <p role="note" className="rounded-2xl border border-destructive/30 bg-destructive/5 p-4 text-sm leading-relaxed">
          {t("writing.result.offTask")}
        </p>
      )}

      {fb.summary_vi && <p className="leading-relaxed">{fb.summary_vi}</p>}

      <ul className="space-y-3">
        {fb.criteria.map((c) => (
          <li key={c.key} className="rounded-2xl border border-border p-4">
            <div className="mb-2 flex items-center justify-between gap-3">
              <h3 className="font-medium">{t(`writing.criteria.${c.key}`)}</h3>
              <span className="font-semibold tabular-nums">{formatBand(c.band)}</span>
            </div>
            <div className="mb-3 h-1.5 overflow-hidden rounded-full bg-muted" aria-hidden>
              <div className="h-full rounded-full bg-primary" style={{ width: `${(c.band / 9) * 100}%` }} />
            </div>
            <p className="text-sm text-muted-foreground">{c.comment_vi}</p>
          </li>
        ))}
      </ul>

      {fb.fixes.length > 0 && (
        <div>
          <h3 className="mb-3 font-medium">{t("writing.result.fixes")}</h3>
          <ol className="space-y-3">
            {fb.fixes.map((f, i) => (
              <li key={i} className="space-y-1.5 rounded-2xl border border-border p-4 text-sm">
                <p className="text-destructive line-through decoration-destructive/40">{f.quote}</p>
                <p className="font-medium text-emerald-700 dark:text-emerald-400">{f.better}</p>
                <p className="text-muted-foreground">{f.explain_vi}</p>
              </li>
            ))}
          </ol>
        </div>
      )}

      <details className="rounded-2xl border border-border p-4" open>
        <summary className="cursor-pointer font-medium">
          {t("writing.result.yourAnswer", { words: s.word_count, time: formatClock(s.seconds_spent) })}
        </summary>
        <p className="mt-3 whitespace-pre-wrap text-sm leading-relaxed">{s.text}</p>
      </details>

      {attempt.model_answer && (
        <details className="rounded-2xl border border-border p-4">
          <summary className="cursor-pointer font-medium">{t("writing.result.modelAnswer")}</summary>
          <p className="mt-3 whitespace-pre-wrap text-sm leading-relaxed">{attempt.model_answer}</p>
        </details>
      )}

      {attempt.vocab_available && (
        <details className="rounded-2xl border border-border p-4">
          <summary className="cursor-pointer font-medium">
            {t("writing.vocab.title")}
            <span className="ml-2 text-sm font-normal text-muted-foreground">{t("writing.vocab.summaryHint")}</span>
          </summary>
          <div className="mt-3">
            <UsefulVocab attempt={attempt} />
          </div>
        </details>
      )}

      <div className="flex flex-col gap-2 border-t border-border pt-4 sm:flex-row sm:justify-end">
        <Button variant="outline" onClick={() => restart(attempt.level)} disabled={again.isPending}>
          {again.isPending ? (
            <LoaderCircle className="size-4 animate-spin" strokeWidth={1.75} aria-hidden />
          ) : (
            <RotateCcw className="size-4" strokeWidth={1.75} aria-hidden />
          )}
          {t("writing.result.again")}
        </Button>
        {attempt.level < 4 && (
          <Button onClick={() => restart((attempt.level + 1) as TLevel)} disabled={again.isPending}>
            {t("writing.result.harder", { level: t(`writing.levels.${attempt.level + 1}.name`) })}
            <ArrowRight className="size-4" strokeWidth={1.75} aria-hidden />
          </Button>
        )}
      </div>
    </section>
  );
}
