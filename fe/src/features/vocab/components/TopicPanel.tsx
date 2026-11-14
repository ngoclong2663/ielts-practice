"use client";

import { Lightbulb, Minus, Plus } from "lucide-react";
import { useTranslations } from "next-intl";
import { useState } from "react";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Skeleton } from "@/components/ui/skeleton";
import { useIsClient } from "@/hooks/use-is-client";
import { ApiError } from "@/lib/api-client";
import { cn } from "@/lib/utils";
import { useCreateSuggestion, useTopics } from "../hooks/use-suggestions";
import { useVocabStatus } from "../hooks/use-vocab";
import { DeckPicker, useTargetDeck } from "./DeckPicker";

const MIN_COUNT = 5;
const MAX_COUNT = 30;
const OTHER = "__other__";

/** Words by topic (Epic 4): topic, count and deck → a mixed list of new words (everyday, IELTS, TOEIC) to review. */
export function TopicPanel({ onCreated }: { onCreated: (suggestionId: string) => void }) {
  const t = useTranslations();
  const isClient = useIsClient();
  const topicsQuery = useTopics();
  const topics = isClient ? topicsQuery.data : undefined;
  const statusQuery = useVocabStatus();
  const status = isClient ? statusQuery.data : undefined;
  const create = useCreateSuggestion();
  const target = useTargetDeck();
  const [topicId, setTopicId] = useState<string | null>(null);
  const [topicText, setTopicText] = useState("");
  const [count, setCount] = useState(15);

  const remaining = status?.suggestions.remaining;
  const hasTopic = topicId === OTHER ? topicText.trim().length > 0 : !!topicId;
  const disabled = !hasTopic || create.isPending || remaining === 0;

  const submit = () => {
    if (disabled) return;
    create.mutate(
      {
        ...(topicId === OTHER ? { topic_text: topicText.trim() } : { topic_id: topicId ?? undefined }),
        count,
        deck_id: target.targetDeckId,
      },
      {
        onSuccess: (s) => onCreated(s.id),
        onError: (error) => {
          const code = error instanceof ApiError ? error.code : "UNKNOWN";
          toast.error(t(`errors.${code}`, { ...(error.params as Record<string, number>) }), { id: "vocab-suggest" });
        },
      },
    );
  };

  return (
    <section className="space-y-6 rounded-2xl border border-border bg-card p-5 sm:p-6">
      <fieldset className="space-y-3">
        <legend className="mb-3 font-medium">{t("vocab.topicLabel")}</legend>
        {!topics ? (
          <Skeleton className="h-20 w-full" />
        ) : (
          <div className="flex flex-wrap gap-2">
            {[...topics, OTHER].map((id) => (
              <button
                key={id}
                type="button"
                aria-pressed={topicId === id}
                onClick={() => setTopicId(id)}
                className={cn(
                  "rounded-full border px-3.5 py-1.5 text-sm transition-colors",
                  topicId === id ? "border-primary bg-primary/10 font-medium text-primary" : "border-border hover:bg-muted",
                )}
              >
                {id === OTHER ? t("vocab.topicOther") : t(`vocab.topics.${id}`)}
              </button>
            ))}
          </div>
        )}
        {topicId === OTHER && (
          <Input
            autoFocus
            value={topicText}
            onChange={(e) => setTopicText(e.target.value)}
            maxLength={60}
            placeholder={t("vocab.topicPlaceholder")}
            aria-label={t("vocab.topicOther")}
            className="sm:w-96"
          />
        )}
      </fieldset>

      <div>
        <div>
          <label htmlFor="topic-count" className="mb-2 block text-sm font-medium">
            {t("vocab.countLabel")}
          </label>
          <div className="flex items-center gap-2">
            <Button
              type="button"
              variant="outline"
              size="icon"
              aria-label={t("vocab.fewer")}
              onClick={() => setCount((c) => Math.max(MIN_COUNT, c - 5))}
              disabled={count <= MIN_COUNT}
            >
              <Minus className="size-4" strokeWidth={1.75} aria-hidden />
            </Button>
            <Input
              id="topic-count"
              type="number"
              inputMode="numeric"
              min={MIN_COUNT}
              max={MAX_COUNT}
              value={count}
              onChange={(e) => setCount(Math.min(MAX_COUNT, Math.max(MIN_COUNT, Number(e.target.value) || MIN_COUNT)))}
              className="w-20 text-center tabular-nums"
            />
            <Button
              type="button"
              variant="outline"
              size="icon"
              aria-label={t("vocab.more")}
              onClick={() => setCount((c) => Math.min(MAX_COUNT, c + 5))}
              disabled={count >= MAX_COUNT}
            >
              <Plus className="size-4" strokeWidth={1.75} aria-hidden />
            </Button>
          </div>
        </div>
      </div>

      <div className="border-t border-border pt-4">
        <DeckPicker id="topic-deck" target={target} />
        <p className="mt-2 text-xs text-muted-foreground">{t("vocab.topicDeckHelp")}</p>
      </div>

      <div className="flex flex-col gap-3 border-t border-border pt-4 sm:flex-row sm:items-center sm:justify-between">
        <p className="text-sm text-muted-foreground">
          {status && t("vocab.suggestionsLeft", { remaining: status.suggestions.remaining, limit: status.suggestions.limit })}
        </p>
        <Button size="lg" onClick={submit} disabled={disabled}>
          <Lightbulb className="size-4" strokeWidth={1.75} aria-hidden />
          {create.isPending ? t("vocab.suggesting") : t("vocab.suggestWords", { count })}
        </Button>
      </div>
    </section>
  );
}
