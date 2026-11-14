"use client";

import { AlertTriangle, Sparkles, X } from "lucide-react";
import { useTranslations } from "next-intl";
import { useMemo, useState } from "react";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import { Textarea } from "@/components/ui/textarea";
import { useIsClient } from "@/hooks/use-is-client";
import { ApiError } from "@/lib/api-client";
import { useCreateJob, useVocabStatus } from "../hooks/use-vocab";
import { lexemeKey, normalizeEntries } from "../lib/normalize";
import { DeckPicker, useTargetDeck } from "./DeckPicker";
import { RecentJobs } from "./RecentJobs";

const MAX_ENTRIES = 30;

export function PasteWordsPanel({ onCreated }: { onCreated: (jobId: string) => void }) {
  const t = useTranslations();
  const isClient = useIsClient();
  const statusQuery = useVocabStatus();
  const status = isClient ? statusQuery.data : undefined;
  const createJob = useCreateJob();
  const [text, setText] = useState("");
  const [dropped, setDropped] = useState<Set<string>>(new Set());
  const target = useTargetDeck();

  const { entries, rejected } = useMemo(() => normalizeEntries(text), [text]);
  const chosen = entries.filter((e) => !dropped.has(lexemeKey(e.text, e.sense_hint)));
  const max = status?.max_entries ?? MAX_ENTRIES;
  const tooMany = chosen.length > max;

  const submit = () => {
    if (chosen.length === 0 || tooMany) return;
    createJob.mutate(
      { entries: chosen, target_deck_id: target.targetDeckId },
      {
        onSuccess: (job) => {
          setText("");
          setDropped(new Set());
          onCreated(job.id);
        },
        onError: (error) => {
          const code = error instanceof ApiError ? error.code : "UNKNOWN";
          toast.error(t(`errors.${code}`, { ...(error.params as Record<string, number>), max }), {
            id: "vocab-create",
          });
        },
      },
    );
  };

  return (
    <div className="space-y-8">
      {status && !status.configured && (
        <div role="status" className="flex items-start gap-3 rounded-2xl border border-srs-learn/40 bg-srs-learn/10 p-4 text-sm">
          <AlertTriangle className="mt-0.5 size-4 shrink-0 text-srs-learn" strokeWidth={1.75} aria-hidden />
          <p>{t("vocab.notConfigured")}</p>
        </div>
      )}

      <section className="space-y-4 rounded-2xl border border-border bg-card p-5 sm:p-6">
        <div className="flex items-baseline justify-between gap-3">
          <label htmlFor="paste-words" className="font-medium">
            {t("vocab.pasteLabel")}
          </label>
          <span className={tooMany ? "text-sm font-medium text-destructive" : "text-sm text-muted-foreground"} aria-live="polite">
            {t("vocab.counter", { count: chosen.length, max })}
          </span>
        </div>
        <Textarea
          id="paste-words"
          rows={6}
          value={text}
          onChange={(e) => setText(e.target.value)}
          placeholder={t("vocab.pastePlaceholder")}
          className="min-h-36 resize-y font-mono text-sm"
          onKeyDown={(e) => {
            if (e.key === "Enter" && (e.metaKey || e.ctrlKey)) submit();
          }}
        />
        <p className="text-xs text-muted-foreground">{t("vocab.pasteHelp")}</p>

        {(entries.length > 0 || rejected.length > 0) && (
          <ul className="flex flex-wrap gap-2" aria-label={t("vocab.previewLabel")}>
            {entries.map((e) => {
              const key = lexemeKey(e.text, e.sense_hint);
              const off = dropped.has(key);
              return (
                <li key={key}>
                  <button
                    type="button"
                    onClick={() =>
                      setDropped((prev) => {
                        const next = new Set(prev);
                        if (off) next.delete(key);
                        else next.add(key);
                        return next;
                      })
                    }
                    aria-pressed={!off}
                    aria-label={off ? t("vocab.includeWord", { word: e.text }) : t("vocab.excludeWord", { word: e.text })}
                    className={
                      off
                        ? "inline-flex items-center gap-1.5 rounded-full border border-dashed border-border px-3 py-1 text-sm text-muted-foreground line-through"
                        : "inline-flex items-center gap-1.5 rounded-full bg-secondary px-3 py-1 text-sm transition-colors hover:bg-secondary/70"
                    }
                  >
                    {e.text}
                    {e.sense_hint && <span className="text-muted-foreground">· {e.sense_hint}</span>}
                    {!off && <X className="size-3.5 text-muted-foreground" strokeWidth={2} aria-hidden />}
                  </button>
                </li>
              );
            })}
            {rejected.map((r) => (
              <li
                key={`rejected-${r.text}`}
                title={t(`vocab.reject.${r.reason}`)}
                className="inline-flex items-center gap-1.5 rounded-full border border-border px-3 py-1 text-sm text-muted-foreground opacity-70"
              >
                <span className="max-w-48 truncate">{r.text}</span>
                <span className="text-xs">({t(`vocab.reject.${r.reason}`)})</span>
              </li>
            ))}
          </ul>
        )}

        <div className="border-t border-border pt-4">
          <DeckPicker id="paste-deck" target={target} />
        </div>

        <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
          <p className="text-sm text-muted-foreground">
            {status && t("vocab.quota", { remaining: status.quota.remaining, limit: status.quota.limit })}
          </p>
          <Button onClick={submit} disabled={chosen.length === 0 || tooMany || createJob.isPending} size="lg">
            <Sparkles className="size-4" strokeWidth={1.75} aria-hidden />
            {createJob.isPending ? t("vocab.generating") : t("vocab.generate", { count: chosen.length })}
          </Button>
        </div>
      </section>

      <RecentJobs onOpen={onCreated} />
    </div>
  );
}
