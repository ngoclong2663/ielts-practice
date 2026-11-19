"use client";

import { ArrowLeft, Plus, Sparkles, Volume2 } from "lucide-react";
import { useTranslations } from "next-intl";
import { useState } from "react";
import { toast } from "sonner";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Checkbox } from "@/components/ui/checkbox";
import { Skeleton } from "@/components/ui/skeleton";
import { useMe } from "@/features/profile/hooks/use-me";
import { useIsClient } from "@/hooks/use-is-client";
import { useRangeSelect } from "@/hooks/use-range-select";
import { useSpeak } from "@/hooks/use-speak";
import { ApiError } from "@/lib/api-client";
import { cn } from "@/lib/utils";
import { useKnownWord, useMoreSuggestions, useSuggestion } from "../hooks/use-suggestions";
import { useCreateJob } from "../hooks/use-vocab";

const MAX_GENERATE = 30;

type SuggestionReviewProps = {
  suggestionId: string;
  onBack: () => void;
  onGenerated: (jobId: string) => void;
};

/** Review suggested topic words: keep, drop, "I know this", suggest more, then generate cards (Epic 4 §3). */
export function SuggestionReview({ suggestionId, onBack, onGenerated }: SuggestionReviewProps) {
  const t = useTranslations();
  const isClient = useIsClient();
  const query = useSuggestion(suggestionId);
  const suggestion = isClient ? query.data : undefined;
  const more = useMoreSuggestions(suggestionId);
  const { mark, unmark } = useKnownWord(suggestionId);
  const createJob = useCreateJob();
  const accent = useMe().data?.study_settings.accent ?? "uk";
  const { speak, supported } = useSpeak(accent === "us" ? "en-US" : "en-GB");
  // Words the user unchecked (everything starts checked).
  const [dropped, setDropped] = useState<Set<string>>(new Set());
  const range = useRangeSelect(suggestion?.words.map((w) => w.lemma) ?? []);

  const onError = (error: ApiError) =>
    toast.error(t(`errors.${error.code}`, { ...(error.params as Record<string, number>) }), { id: "vocab-suggest" });

  const back = (
    <Button variant="ghost" size="sm" onClick={onBack} className="-ml-2 mb-4">
      <ArrowLeft className="size-4" strokeWidth={1.75} aria-hidden />
      {t("vocab.newTopic")}
    </Button>
  );

  if (query.error) {
    const code = query.error instanceof ApiError ? query.error.code : "UNKNOWN";
    return (
      <div className="mx-auto max-w-3xl">
        {back}
        <p className="rounded-2xl border border-border bg-card p-6 text-muted-foreground">{t(`errors.${code}`)}</p>
      </div>
    );
  }
  if (!suggestion) {
    return (
      <div className="mx-auto max-w-3xl space-y-3">
        {back}
        <Skeleton className="h-10 w-2/3" />
        {Array.from({ length: 6 }, (_, i) => (
          <Skeleton key={i} className="h-14 w-full rounded-xl" />
        ))}
      </div>
    );
  }

  const topic = suggestion.topic_id ? t(`vocab.topics.${suggestion.topic_id}`) : suggestion.topic_text;
  const kept = suggestion.words.filter((w) => !dropped.has(w.lemma));
  // With Shift, every word between the last tick and this one takes its new state (kept or dropped).
  const tick = (lemma: string) => {
    const keep = dropped.has(lemma);
    const keys = range.affected(lemma);
    setDropped((prev) => {
      const next = new Set(prev);
      for (const k of keys) {
        if (keep) next.delete(k);
        else next.add(k);
      }
      return next;
    });
  };
  const markKnown = (lemma: string) =>
    mark.mutate(lemma, {
      onSuccess: () =>
        toast(t("vocab.knownToast", { word: lemma }), {
          action: { label: t("vocab.undoRemove"), onClick: () => unmark.mutate(lemma) },
        }),
      onError,
    });

  const generate = () => {
    createJob.mutate(
      {
        entries: kept.slice(0, MAX_GENERATE).map((w) => ({ text: w.lemma, sense_hint: "" })),
        target_deck_id: suggestion.deck_id ?? undefined,
      },
      { onSuccess: (job) => onGenerated(job.id), onError },
    );
  };

  return (
    <div className="mx-auto max-w-3xl">
      {back}
      <header className="mb-5 space-y-1">
        <h1 className="text-2xl font-semibold tracking-tight">{t("vocab.suggestionTitle", { topic })}</h1>
        <p className="text-sm text-muted-foreground">
          {t("vocab.wordCount", { count: suggestion.words.length })}
          {suggestion.skipped_existing > 0 && <> · {t("vocab.skippedExisting", { count: suggestion.skipped_existing })}</>}
        </p>
        <p className="text-sm text-muted-foreground">{t("vocab.suggestionHint")}</p>
      </header>

      {suggestion.words.length === 0 ? (
        <p className="rounded-2xl border border-dashed border-border p-6 text-center text-muted-foreground">
          {t("vocab.noNewWords")}
        </p>
      ) : (
        <ul className="divide-y divide-border overflow-hidden rounded-2xl border border-border bg-card">
          {suggestion.words.map((w) => {
            const checked = !dropped.has(w.lemma);
            const id = `word-${w.lemma.replace(/\s+/g, "-")}`;
            return (
              <li key={w.lemma} {...range.rowProps} className={cn("flex items-center gap-3 px-4 py-3", !checked && "opacity-55")}>
                <Checkbox id={id} checked={checked} onCheckedChange={() => tick(w.lemma)} aria-label={t("vocab.keepWord", { word: w.lemma })} />
                <label htmlFor={id} className="min-w-0 flex-1 cursor-pointer">
                  <span className="font-medium">{w.lemma}</span>
                  {w.pos && <span className="ml-1.5 text-sm italic text-muted-foreground">{w.pos}</span>}
                  {w.cefr && (
                    <Badge variant="secondary" className="ml-2 align-middle tabular-nums">
                      {w.cefr}
                    </Badge>
                  )}
                  <span className="block text-sm text-primary">{w.meaning_vi}</span>
                </label>
                {supported && (
                  <Button variant="ghost" size="icon-sm" onClick={() => speak(w.lemma)} aria-label={t("vocab.playWord", { word: w.lemma })}>
                    <Volume2 className="size-4" strokeWidth={1.75} aria-hidden />
                  </Button>
                )}
                <Button variant="ghost" size="sm" className="shrink-0 text-muted-foreground" onClick={() => markKnown(w.lemma)}>
                  {t("vocab.iKnowThis")}
                </Button>
              </li>
            );
          })}
        </ul>
      )}

      <div className="sticky bottom-[calc(4.25rem+env(safe-area-inset-bottom))] z-20 mt-6 lg:bottom-6">
        <div className="flex flex-wrap items-center justify-between gap-x-3 gap-y-2 rounded-2xl border border-border bg-background/95 px-4 py-3 shadow-lg backdrop-blur supports-[backdrop-filter]:bg-background/85">
          <div className="flex items-center gap-3">
            <Button
              variant="outline"
              onClick={() => more.mutate(suggestion.requested, { onError })}
              disabled={more.isPending || suggestion.remaining_requests === 0}
            >
              <Plus className="size-4" strokeWidth={1.75} aria-hidden />
              {more.isPending ? t("vocab.suggesting") : t("vocab.suggestMore")}
            </Button>
            <span className="hidden text-xs text-muted-foreground sm:inline">
              {t("vocab.requestsLeft", { count: suggestion.remaining_requests })}
            </span>
          </div>
          <Button size="lg" onClick={generate} disabled={kept.length === 0 || createJob.isPending}>
            <Sparkles className="size-4" strokeWidth={1.75} aria-hidden />
            {createJob.isPending ? t("vocab.generating") : t("vocab.generate", { count: Math.min(kept.length, MAX_GENERATE) })}
          </Button>
          <p className="w-full text-xs text-muted-foreground sm:hidden">
            {t("vocab.requestsLeft", { count: suggestion.remaining_requests })}
          </p>
        </div>
        {kept.length > MAX_GENERATE && (
          <p className="mt-2 text-center text-xs text-muted-foreground">{t("vocab.tooManyKept", { max: MAX_GENERATE })}</p>
        )}
      </div>
    </div>
  );
}
