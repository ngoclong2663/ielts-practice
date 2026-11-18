"use client";

import { BookPlus, ChevronDown, Lightbulb, LoaderCircle, RotateCcw, Volume2 } from "lucide-react";
import { useTranslations } from "next-intl";
import { Button } from "@/components/ui/button";
import { Checkbox } from "@/components/ui/checkbox";
import { Skeleton } from "@/components/ui/skeleton";
import { useSpeak } from "@/hooks/use-speak";
import { cn } from "@/lib/utils";
import { useAttemptVocab } from "../hooks/use-writing";
import { useVocabSelection } from "../hooks/use-vocab-selection";
import type { TAttempt, TVocabGroup, TVocabItem } from "../types";

type UsefulVocabProps = {
  attempt: TAttempt;
  /** Compact: terms and synonyms only, no selection (a reminder next to the answer box). */
  compact?: boolean;
};

/** Epic 7: the lesson's useful vocabulary (topic words, paraphrases, task language) with synonyms. */
export function UsefulVocab({ attempt, compact = false }: UsefulVocabProps) {
  const t = useTranslations("writing.vocab");
  const query = useAttemptVocab(attempt.id, attempt.vocab_available);

  if (!attempt.vocab_available) return null;
  if (query.isPending) {
    return (
      <div className="space-y-3" aria-busy="true">
        <p className="flex items-center gap-2 text-sm text-muted-foreground">
          <LoaderCircle className="size-4 animate-spin" strokeWidth={1.75} aria-hidden />
          {t("loading")}
        </p>
        <Skeleton className="h-24 rounded-2xl" />
        <Skeleton className="h-24 rounded-2xl" />
      </div>
    );
  }
  if (query.isError) {
    return (
      <div className="flex flex-wrap items-center justify-between gap-3 rounded-2xl border border-border p-4 text-sm">
        <p className="text-muted-foreground">{t("loadError")}</p>
        <Button size="sm" variant="outline" onClick={() => query.refetch()}>
          <RotateCcw className="size-4" strokeWidth={1.75} aria-hidden />
          {t("retry")}
        </Button>
      </div>
    );
  }
  const groups = query.data.groups;
  return compact ? <CompactVocab groups={groups} /> : <FullVocab groups={sortForWriting(groups)} taskNo={attempt.task.task_no} />;
}

// Groups in the order the answer uses them: the introduction paraphrases the question, then the overview and
// comparisons use the task language, and the body paragraphs carry the topic words.
const WRITING_ORDER = ["paraphrase", "task", "topic"];
const sortForWriting = (groups: TVocabGroup[]) =>
  [...groups].sort((a, b) => WRITING_ORDER.indexOf(a.key) - WRITING_ORDER.indexOf(b.key));

function FullVocab({ groups, taskNo }: { groups: TVocabGroup[]; taskNo: number }) {
  const t = useTranslations("writing.vocab");
  const items = groups.flatMap((g) => g.items);
  const sel = useVocabSelection(items);

  return (
    <div>
      <p className="mb-2 text-sm text-muted-foreground">{t("intro")}</p>
      {/* One fold per group (the first open), so the list stays short and each part is found where it is used.
          Plain rows with dividers, not boxes: the section around them is already a card. */}
      {groups.map((g, i) => {
        const picked = g.items.filter((it) => sel.selected.has(it.term)).length;
        return (
          <details key={g.key} open={i === 0 || undefined} className="group/vocab @container border-b border-border last-of-type:border-b-0">
            <summary className="flex cursor-pointer list-none items-center gap-3 py-3 [&::-webkit-details-marker]:hidden hover:[&_.vocab-title]:text-primary">
              <span className="min-w-0 flex-1">
                <span className="flex flex-wrap items-center gap-x-2 gap-y-1">
                  <span className="vocab-title font-medium transition-colors">{t(`groups.${g.key}.title`)}</span>
                  <span className="rounded-full bg-primary/10 px-2 py-0.5 text-xs font-medium text-primary">
                    {t(`usedIn.${g.key === "task" ? `task${taskNo === 2 ? 2 : 1}` : g.key}`)}
                  </span>
                </span>
                <span className="mt-0.5 block text-sm text-muted-foreground">{t(`groups.${g.key}.help`)}</span>
              </span>
              <span className="shrink-0 text-right text-xs text-muted-foreground tabular-nums">
                {t("wordCount", { count: g.items.length })}
                {picked > 0 && <span className="block font-medium text-primary">{t("pickedCount", { count: picked })}</span>}
              </span>
              <ChevronDown className="size-4 shrink-0 text-muted-foreground transition-transform group-open/vocab:rotate-180" strokeWidth={1.75} aria-hidden />
            </summary>
            {/* Two cards a row only where the fold itself is wide; the attempt column is narrow on desktop. */}
            <ul className="grid gap-3 pb-4 @2xl:grid-cols-2">
              {g.items.map((item) => (
                <VocabCard key={item.term} item={item} checked={sel.selected.has(item.term)} onToggle={() => sel.toggle(item.term)} />
              ))}
            </ul>
          </details>
        );
      })}

      <div className="sticky bottom-3 z-10 mt-3 flex flex-col gap-3 rounded-2xl border border-border bg-background/95 p-3 shadow-sm backdrop-blur sm:flex-row sm:items-center sm:justify-between">
        <label className="flex cursor-pointer items-center gap-2 text-sm">
          <Checkbox checked={sel.allSelected} onCheckedChange={sel.toggleAll} />
          {t("selectAll", { count: items.length })}
        </label>
        <Button onClick={sel.save} disabled={sel.selected.size === 0 || sel.saving}>
          {sel.saving ? (
            <LoaderCircle className="size-4 animate-spin" strokeWidth={1.75} aria-hidden />
          ) : (
            <BookPlus className="size-4" strokeWidth={1.75} aria-hidden />
          )}
          {t("addToDeck", { count: sel.selected.size })}
        </Button>
      </div>
    </div>
  );
}

type VocabCardProps = { item: TVocabItem; checked: boolean; onToggle: () => void };

function VocabCard({ item, checked, onToggle }: VocabCardProps) {
  const t = useTranslations("writing.vocab");
  const { speak, supported } = useSpeak("en-GB");
  const id = `vocab-item-${item.term.replace(/\W+/g, "-")}`;

  return (
    <li
      className={cn(
        "flex gap-3 rounded-2xl border p-4 transition-colors",
        checked ? "border-primary/50 bg-primary/5" : "border-border",
      )}
    >
      <Checkbox id={id} checked={checked} onCheckedChange={onToggle} className="mt-1" aria-label={t("select", { term: item.term })} />
      <div className="min-w-0 flex-1 space-y-1.5">
        <div className="flex items-start justify-between gap-2">
          <label htmlFor={id} className="cursor-pointer">
            <span className="font-semibold" lang="en">
              {item.term}
            </span>
            <span className="ml-2 text-xs text-muted-foreground">{t(`pos.${item.pos}`)}</span>
            <span className="block text-sm text-muted-foreground">{item.vi}</span>
          </label>
          {supported && (
            <Button
              size="icon"
              variant="ghost"
              className="size-8 shrink-0"
              onClick={() => speak(item.term)}
              aria-label={t("listen", { term: item.term })}
            >
              <Volume2 className="size-4" strokeWidth={1.75} aria-hidden />
            </Button>
          )}
        </div>
        {item.synonyms.length > 0 && (
          <p className="text-sm">
            <span className="text-muted-foreground">{t("synonyms")} </span>
            <span lang="en" className="font-medium text-primary">
              {item.synonyms.join(" · ")}
            </span>
          </p>
        )}
        {item.example_en && (
          <p className="text-sm italic text-muted-foreground" lang="en">
            {item.example_en}
          </p>
        )}
        {item.note_vi && (
          <p className="flex gap-1.5 text-xs text-muted-foreground">
            <Lightbulb className="mt-0.5 size-3.5 shrink-0 text-srs-learn" strokeWidth={1.75} aria-hidden />
            {item.note_vi}
          </p>
        )}
      </div>
    </li>
  );
}

function CompactVocab({ groups }: { groups: TVocabGroup[] }) {
  const t = useTranslations("writing.vocab");
  return (
    <div className="space-y-3">
      {groups.map((g) => (
        <div key={g.key}>
          <h4 className="mb-1.5 text-xs font-medium uppercase tracking-wide text-muted-foreground">{t(`groups.${g.key}.title`)}</h4>
          <ul className="flex flex-wrap gap-2">
            {g.items.map((item) => (
              <li key={item.term} className="rounded-full bg-muted px-3 py-1 text-xs" lang="en">
                <span className="font-medium">{item.term}</span>
                {item.synonyms.length > 0 && <span className="text-muted-foreground"> ≈ {item.synonyms.join(", ")}</span>}
              </li>
            ))}
          </ul>
        </div>
      ))}
    </div>
  );
}
