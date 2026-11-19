"use client";

import { ArrowLeft, Search, X } from "lucide-react";
import { useLocale, useTranslations } from "next-intl";
import { useParams, useSearchParams } from "next/navigation";
import { useMemo, useState } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Skeleton } from "@/components/ui/skeleton";
import { useMe } from "@/features/profile/hooks/use-me";
import { useDebouncedValue } from "@/hooks/use-debounced-value";
import { useRangeSelect } from "@/hooks/use-range-select";
import { useIsClient } from "@/hooks/use-is-client";
import { useSpeak } from "@/hooks/use-speak";
import { Link, usePathname, useRouter } from "@/i18n/navigation";
import { ApiError } from "@/lib/api-client";
import { cn } from "@/lib/utils";
import { useCollections, useLibraryWords } from "../hooks/use-library";
import type { TLibraryLevel, TLibraryWord } from "../types";
import { AddToDeckBar } from "./AddToDeckBar";
import { LibraryWordRow } from "./LibraryWordRow";
import { WordPreviewDialog } from "./WordPreviewDialog";

const LEVELS: TLibraryLevel[] = ["A1", "A2", "B1", "B2", "C1", "C2"];

/**
 * One collection (Epic 8 §3.5): topics, a level filter and a search across topics; tick words (across topics too)
 * and add them to a deck at once.
 */
export function CollectionView() {
  const slug = String(useParams<{ slug: string }>().slug ?? "");
  const t = useTranslations();
  const locale = useLocale() as "vi" | "en";
  const isClient = useIsClient();
  const params = useSearchParams();
  const router = useRouter();
  const pathname = usePathname();
  const accent = useMe().data?.study_settings.accent ?? "uk";
  const { speak, supported } = useSpeak(accent === "us" ? "en-US" : "en-GB");

  const [search, setSearch] = useState("");
  const query = useDebouncedValue(search.trim(), 300);
  const searching = query.length >= 2;
  const [level, setLevel] = useState<TLibraryLevel | null>(null);
  const [selected, setSelected] = useState<Map<string, TLibraryWord>>(new Map());
  const [preview, setPreview] = useState<TLibraryWord | null>(null);

  const collections = useCollections();
  const collection = isClient ? collections.data?.find((c) => c.slug === slug) : undefined;
  const urlTopic = params.get("topic");
  const topic = collection?.topics.some((tp) => tp.id === urlTopic) ? urlTopic! : collection?.topics[0]?.id;
  const topicQuery = useLibraryWords(slug, { topic }, !!topic && !searching);
  const searchQuery = useLibraryWords(slug, { q: query }, searching);
  const active = searching ? searchQuery : topicQuery;
  const data = isClient ? active.data : undefined;

  const words = useMemo(() => (data?.words ?? []).filter((w) => !level || w.cefr === level), [data, level]);
  const levels = useMemo(() => LEVELS.filter((l) => data?.words.some((w) => w.cefr === l)), [data]);
  const selectable = words.filter((w) => !w.in_deck);
  const range = useRangeSelect(words.map((w) => w.lexeme_id));
  const allSelected = selectable.length > 0 && selectable.every((w) => selected.has(w.lexeme_id));

  const setTopic = (id: string) => {
    setLevel(null);
    router.replace(`${pathname}?topic=${encodeURIComponent(id)}`, { scroll: false });
  };
  const toggle = (w: TLibraryWord) =>
    setSelected((prev) => {
      const next = new Map(prev);
      if (next.has(w.lexeme_id)) next.delete(w.lexeme_id);
      else next.set(w.lexeme_id, w);
      return next;
    });
  // A tick in the list: with Shift, the whole range from the last tick takes the new state.
  const tick = (w: TLibraryWord) => {
    const on = !selected.has(w.lexeme_id);
    const keys = range.affected(w.lexeme_id);
    setSelected((prev) => {
      const next = new Map(prev);
      for (const x of words) {
        if (!keys.includes(x.lexeme_id) || x.in_deck) continue;
        if (on) next.set(x.lexeme_id, x);
        else next.delete(x.lexeme_id);
      }
      return next;
    });
  };
  const toggleAll = () =>
    setSelected((prev) => {
      const next = new Map(prev);
      for (const w of selectable) {
        if (allSelected) next.delete(w.lexeme_id);
        else next.set(w.lexeme_id, w);
      }
      return next;
    });

  const back = (
    <Button variant="ghost" size="sm" className="-ml-2 mb-4" nativeButton={false} render={<Link href="/library" />}>
      <ArrowLeft className="size-4" strokeWidth={1.75} aria-hidden />
      {t("library.title")}
    </Button>
  );

  const missing = collections.isSuccess && !collections.data.some((c) => c.slug === slug);
  if (collections.error || missing) {
    const code = collections.error instanceof ApiError ? collections.error.code : "LIBRARY_NOT_FOUND";
    return (
      <div className="mx-auto max-w-3xl">
        {back}
        <p className="rounded-2xl border border-border bg-card p-6 text-muted-foreground">{t(`errors.${code}`)}</p>
      </div>
    );
  }
  if (!collection) {
    return (
      <div className="mx-auto max-w-3xl space-y-3">
        {back}
        <Skeleton className="h-9 w-1/2" />
        <Skeleton className="h-10 w-full rounded-xl" />
        {Array.from({ length: 6 }, (_, i) => (
          <Skeleton key={i} className="h-16 w-full rounded-xl" />
        ))}
      </div>
    );
  }

  const selectedTopics = new Set([...selected.values()].map((w) => w.topic));
  const deckName =
    selectedTopics.size === 1
      ? `${collection.title[locale]} · ${t(`vocab.topics.${[...selectedTopics][0]}`)}`
      : collection.title[locale];

  return (
    <div className="mx-auto max-w-3xl">
      {back}
      <header className="mb-5 space-y-1">
        <h1 className="text-2xl font-semibold tracking-tight sm:text-3xl">{collection.title[locale]}</h1>
        <p className="text-sm text-muted-foreground">
          {t("library.counts", { words: collection.words, topics: collection.topics.length })}
        </p>
        <p className="text-sm text-muted-foreground">{t("library.collectionHint")}</p>
      </header>

      <div className="relative mb-4">
        <Search className="pointer-events-none absolute top-1/2 left-3 size-4 -translate-y-1/2 text-muted-foreground" strokeWidth={1.75} aria-hidden />
        <Input
          type="search"
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          placeholder={t("library.searchPlaceholder")}
          aria-label={t("library.searchLabel")}
          maxLength={60}
          className="h-10 pl-9"
        />
        {search && (
          <Button variant="ghost" size="icon-sm" className="absolute top-1/2 right-1.5 -translate-y-1/2" onClick={() => setSearch("")} aria-label={t("library.clearSearch")}>
            <X className="size-4" strokeWidth={1.75} aria-hidden />
          </Button>
        )}
      </div>

      {!searching && (
        <div role="tablist" aria-label={t("library.topics")} className="-mx-4 mb-3 flex gap-2 overflow-x-auto px-4 pb-1 sm:mx-0 sm:flex-wrap sm:px-0">
          {collection.topics.map((tp) => (
            <button
              key={tp.id}
              type="button"
              role="tab"
              aria-selected={tp.id === topic}
              onClick={() => setTopic(tp.id)}
              className={cn(
                "shrink-0 rounded-full border px-3 py-1 text-sm whitespace-nowrap transition-colors",
                tp.id === topic ? "border-primary bg-primary text-primary-foreground" : "border-border text-muted-foreground hover:text-foreground",
              )}
            >
              {t(`vocab.topics.${tp.id}`)} <span className="tabular-nums opacity-70">{tp.words}</span>
            </button>
          ))}
        </div>
      )}

      <div className="mb-2 flex flex-wrap items-center justify-between gap-2">
        <div className="flex items-center gap-1" role="group" aria-label={t("library.level")}>
          {levels.length > 1 &&
            [null, ...levels].map((l) => (
              <button
                key={l ?? "all"}
                type="button"
                aria-pressed={level === l}
                onClick={() => setLevel(l)}
                className={cn(
                  "rounded-md px-2 py-0.5 text-xs font-medium tabular-nums transition-colors",
                  level === l ? "bg-secondary text-secondary-foreground" : "text-muted-foreground hover:text-foreground",
                )}
              >
                {l ?? t("library.allLevels")}
              </button>
            ))}
        </div>
        {selectable.length > 0 && (
          <Button variant="ghost" size="sm" onClick={toggleAll}>
            {allSelected ? t("library.unselectAll") : t("library.selectAll", { count: selectable.length })}
          </Button>
        )}
      </div>

      {searching && (
        <p className="mb-2 text-sm text-muted-foreground" aria-live="polite">
          {searchQuery.isFetching && !searchQuery.data ? t("common.loading") : t("library.results", { count: words.length, query })}
        </p>
      )}

      {!data ? (
        <div className="space-y-2">
          {Array.from({ length: 6 }, (_, i) => (
            <Skeleton key={i} className="h-16 w-full rounded-xl" />
          ))}
        </div>
      ) : words.length === 0 ? (
        <p className="rounded-2xl border border-dashed border-border p-6 text-center text-muted-foreground">
          {searching ? t("library.noResults") : t("library.noWords")}
        </p>
      ) : (
        <ul className={cn("divide-y divide-border overflow-hidden rounded-2xl border border-border bg-card transition-opacity", active.isFetching && "opacity-70")}>
          {words.map((w) => (
            <LibraryWordRow
              key={w.lexeme_id}
              word={w}
              accent={accent}
              selected={selected.has(w.lexeme_id)}
              showTopic={searching}
              onToggle={() => tick(w)}
              rowProps={range.rowProps}
              onPreview={() => setPreview(w)}
              onSpeak={supported ? () => speak(w.card.headword) : undefined}
            />
          ))}
        </ul>
      )}
      {data?.truncated && <p className="mt-2 text-center text-xs text-muted-foreground">{t("library.truncated")}</p>}

      <AddToDeckBar
        lexemeIds={[...selected.keys()]}
        defaultDeckName={deckName}
        onAdded={() => setSelected(new Map())}
        onClear={() => setSelected(new Map())}
      />

      <WordPreviewDialog
        word={preview}
        accent={accent}
        selected={!!preview && selected.has(preview.lexeme_id)}
        onToggle={() => preview && toggle(preview)}
        onOpenChange={(open) => !open && setPreview(null)}
      />
    </div>
  );
}
