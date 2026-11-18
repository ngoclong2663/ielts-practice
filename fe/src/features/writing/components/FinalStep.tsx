"use client";

import { ClipboardPaste, LoaderCircle, Send, Timer } from "lucide-react";
import { useTranslations } from "next-intl";
import { useEffect, useState } from "react";
import { Button } from "@/components/ui/button";
import { Textarea } from "@/components/ui/textarea";
import { cn } from "@/lib/utils";
import { useSubmitAttempt, useWritingStatus } from "../hooks/use-writing";
import { clearDraft, useDraft } from "../lib/drafts";
import { toastApiError } from "../lib/errors";
import { countWords, formatClock } from "../lib/format";
import type { TAttempt } from "../types";
import { UsefulVocab } from "./UsefulVocab";

const MIN_SUBMIT_WORDS = 20;

/** The learner's checked sentences (corrected where the coach fixed them), one paragraph per block. */
function draftFromSentences(a: TAttempt): string {
  return a.paragraphs
    .filter((p) => p.checked_count > 0)
    .map((p) =>
      p.answers
        .map((ans, i) => {
          const fb = p.feedback[i];
          if (!fb || fb.verdict === "empty") return "";
          return fb.verdict === "correct" ? ans : fb.corrected_en || ans;
        })
        .filter(Boolean)
        .join(" "),
    )
    .filter(Boolean)
    .join("\n\n");
}

export function FinalStep({ attempt, onSubmitted }: { attempt: TAttempt; onSubmitted: () => void }) {
  const t = useTranslations();
  const submit = useSubmitAttempt(attempt.id);
  const status = useWritingStatus().data;
  const draftKey = `${attempt.id}.final`;
  const [text, setText] = useDraft(draftKey, "");
  const [startedAt] = useState(() => Date.now());
  const [now, setNow] = useState(startedAt);
  const words = countWords(text);
  const task = attempt.task;
  const sentences = attempt.level <= 2 ? draftFromSentences(attempt) : "";

  useEffect(() => {
    const id = window.setInterval(() => setNow(Date.now()), 1000);
    return () => window.clearInterval(id);
  }, []);
  const elapsed = (now - startedAt) / 1000;

  const send = () =>
    submit.mutate(
      { text, seconds_spent: Math.round(elapsed) },
      {
        onSuccess: () => {
          clearDraft(draftKey);
          for (const p of attempt.paragraphs) clearDraft(`${attempt.id}.${p.key}`);
          onSubmitted();
        },
        onError: (e) => toastApiError(t, e, "writing-submit"),
      },
    );

  return (
    <section className="space-y-5">
      <header>
        <h2 className="text-lg font-semibold">{t("writing.final.title")}</h2>
        <p className="mt-1 text-sm text-muted-foreground">{t(`writing.final.help${attempt.level <= 2 ? "Guided" : "Free"}`)}</p>
      </header>

      {sentences && (
        <div className="rounded-2xl border border-border bg-muted/30 p-4">
          <div className="mb-2 flex flex-wrap items-center justify-between gap-2">
            <h3 className="text-sm font-medium">{t("writing.final.yourSentences")}</h3>
            <Button
              size="sm"
              variant="outline"
              onClick={() => {
                if (!text.trim() || window.confirm(t("writing.final.replaceConfirm"))) setText(sentences);
              }}
            >
              <ClipboardPaste className="size-4" strokeWidth={1.75} aria-hidden />
              {t("writing.final.useSentences")}
            </Button>
          </div>
          <p className="whitespace-pre-wrap text-sm leading-relaxed text-muted-foreground">{sentences}</p>
        </div>
      )}

      {attempt.vocab_available && (
        <details className="rounded-2xl border border-border p-4">
          <summary className="cursor-pointer text-sm font-medium">{t("writing.vocab.reminder")}</summary>
          <div className="mt-3">
            <UsefulVocab attempt={attempt} compact />
          </div>
        </details>
      )}

      <div>
        <Textarea
          value={text}
          onChange={(e) => setText(e.target.value)}
          placeholder={t("writing.final.placeholder")}
          aria-label={t("writing.final.title")}
          className="min-h-[22rem] text-[15px] leading-relaxed"
          maxLength={6000}
          lang="en"
          spellCheck={false}
        />
        <div className="mt-2 flex flex-wrap items-center justify-between gap-2 text-sm">
          <span className={cn("tabular-nums", words < task.min_words ? "text-amber-600 dark:text-amber-400" : "text-emerald-600 dark:text-emerald-400")}>
            {t("writing.final.wordCount", { words, min: task.min_words })}
          </span>
          <span className={cn("inline-flex items-center gap-1.5 tabular-nums text-muted-foreground", elapsed > task.minutes * 60 && "text-amber-600")}>
            <Timer className="size-4" strokeWidth={1.75} aria-hidden />
            {formatClock(elapsed)} / {formatClock(task.minutes * 60)}
          </span>
        </div>
      </div>

      {words > 0 && words < task.min_words && (
        <p className="text-sm text-amber-700 dark:text-amber-400">{t("writing.final.underLength", { min: task.min_words })}</p>
      )}

      <div className="flex flex-col gap-3 border-t border-border pt-4 sm:flex-row sm:items-center sm:justify-between">
        <p className="text-sm text-muted-foreground">
          {status && t("writing.submissionsLeft", { remaining: status.submissions.remaining, limit: status.submissions.limit })}
        </p>
        <Button size="lg" onClick={send} disabled={words < MIN_SUBMIT_WORDS || submit.isPending || status?.submissions.remaining === 0}>
          {submit.isPending ? (
            <LoaderCircle className="size-4 animate-spin" strokeWidth={1.75} aria-hidden />
          ) : (
            <Send className="size-4" strokeWidth={1.75} aria-hidden />
          )}
          {submit.isPending ? t("writing.final.marking") : t("writing.final.submit")}
        </Button>
      </div>
    </section>
  );
}
