"use client";

import { ArrowRight, Lightbulb, LoaderCircle, SpellCheck } from "lucide-react";
import { useTranslations } from "next-intl";
import { useState } from "react";
import { Button } from "@/components/ui/button";
import { Textarea } from "@/components/ui/textarea";
import { cn } from "@/lib/utils";
import { useCheckParagraph, useWritingStatus } from "../hooks/use-writing";
import { useDraft } from "../lib/drafts";
import { toastApiError } from "../lib/errors";
import type { TAttempt, TParagraph } from "../types";
import { SentenceFeedback } from "./SentenceFeedback";

/** One paragraph, sentence by sentence: the prompt at the learner's level, hints, and feedback after a check. */
export function ParagraphStep({ attempt, paragraph, onNext }: { attempt: TAttempt; paragraph: TParagraph; onNext: () => void }) {
  const t = useTranslations();
  const check = useCheckParagraph(attempt.id);
  const status = useWritingStatus().data;
  const empty = paragraph.sentences.map(() => "");
  const [answers, setAnswers] = useDraft<string[]>(
    `${attempt.id}.${paragraph.key}`,
    paragraph.answers.length === paragraph.sentences.length ? paragraph.answers : empty,
  );
  const [openHints, setOpenHints] = useState<Set<number>>(new Set());
  const checked = paragraph.checked_count > 0;
  const changed = checked && answers.some((a, i) => a.trim() !== (paragraph.answers[i] ?? ""));
  const hasText = answers.some((a) => a.trim() !== "");

  const submit = () =>
    check.mutate(
      { key: paragraph.key, answers: answers.map((a) => a.trim()) },
      { onError: (e) => toastApiError(t, e, "writing-check") },
    );

  const toggleHints = (i: number) =>
    setOpenHints((s) => {
      const next = new Set(s);
      if (next.has(i)) next.delete(i);
      else next.add(i);
      return next;
    });

  return (
    <section className="space-y-5">
      <header>
        <h2 className="text-lg font-semibold">{t(`writing.stepNames.${paragraph.key}`)}</h2>
        <p className="mt-1 text-sm text-muted-foreground">{paragraph.goal_vi}</p>
        <p className="mt-1 text-xs text-muted-foreground">{t(`writing.promptHelp.${attempt.level}`)}</p>
      </header>

      <ol className="space-y-4">
        {paragraph.sentences.map((s, i) => (
          <li key={i} className="space-y-2 rounded-2xl border border-border p-4">
            <div className="flex items-start gap-3">
              <span className="mt-0.5 flex size-6 shrink-0 items-center justify-center rounded-full bg-muted text-xs font-medium tabular-nums">
                {i + 1}
              </span>
              <p className={cn("flex-1 text-[15px] leading-relaxed", attempt.level === 3 && "font-mono text-sm")}>
                {s.prompt}
              </p>
              {s.hints.length > 0 && (
                <Button
                  type="button"
                  variant="ghost"
                  size="icon"
                  aria-expanded={openHints.has(i)}
                  aria-label={t("writing.hints")}
                  onClick={() => toggleHints(i)}
                  className="-mt-1 shrink-0"
                >
                  <Lightbulb className="size-4" strokeWidth={1.75} aria-hidden />
                </Button>
              )}
            </div>
            {openHints.has(i) && (
              <ul className="ml-9 flex flex-wrap gap-2">
                {s.hints.map((h) => (
                  <li key={h.en} className="rounded-full bg-amber-500/10 px-3 py-1 text-xs">
                    <span className="font-medium">{h.en}</span>
                    <span className="text-muted-foreground"> · {h.vi}</span>
                  </li>
                ))}
              </ul>
            )}
            <div className="ml-9 space-y-2">
              <Textarea
                value={answers[i] ?? ""}
                onChange={(e) => setAnswers(answers.map((a, j) => (j === i ? e.target.value : a)))}
                placeholder={t("writing.sentencePlaceholder")}
                aria-label={t("writing.sentenceLabel", { n: i + 1 })}
                maxLength={400}
                rows={2}
                lang="en"
                spellCheck={false}
                autoCapitalize="sentences"
              />
              {checked && paragraph.feedback[i] && !changed && (
                <SentenceFeedback feedback={paragraph.feedback[i]} answer={paragraph.answers[i] ?? ""} model={s.model_en} />
              )}
            </div>
          </li>
        ))}
      </ol>

      <div className="flex flex-col gap-3 border-t border-border pt-4 sm:flex-row sm:items-center sm:justify-between">
        <p className="text-sm text-muted-foreground">
          {status && t("writing.checksLeft", { remaining: status.checks.remaining, limit: status.checks.limit })}
        </p>
        <div className="flex gap-2">
          {(!checked || changed) && (
            <Button onClick={submit} disabled={!hasText || check.isPending || status?.checks.remaining === 0}>
              {check.isPending ? (
                <LoaderCircle className="size-4 animate-spin" strokeWidth={1.75} aria-hidden />
              ) : (
                <SpellCheck className="size-4" strokeWidth={1.75} aria-hidden />
              )}
              {check.isPending ? t("writing.checking") : checked ? t("writing.checkAgain") : t("writing.checkParagraph")}
            </Button>
          )}
          <Button variant={checked && !changed ? "default" : "outline"} onClick={onNext}>
            {checked ? t("writing.nextStep") : t("writing.skip")}
            <ArrowRight className="size-4" strokeWidth={1.75} aria-hidden />
          </Button>
        </div>
      </div>
    </section>
  );
}
