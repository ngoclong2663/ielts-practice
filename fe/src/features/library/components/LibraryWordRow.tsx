"use client";

import { Check, Eye, Volume2 } from "lucide-react";
import { useTranslations } from "next-intl";
import { useState } from "react";
import type { HTMLAttributes } from "react";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Checkbox } from "@/components/ui/checkbox";
import { cn } from "@/lib/utils";
import type { TLibraryWord } from "../types";

type LibraryWordRowProps = {
  word: TLibraryWord;
  selected: boolean;
  accent: "uk" | "us";
  showTopic: boolean;
  onToggle: () => void;
  onPreview: () => void;
  onSpeak?: () => void;
  /** Shift-click range selection (useRangeSelect). */
  rowProps?: HTMLAttributes<HTMLLIElement>;
};

/** One library word: tick to select, open the full card, listen. A word already in a deck cannot be selected. */
export function LibraryWordRow({ word, selected, accent, showTopic, onToggle, onPreview, onSpeak, rowProps }: LibraryWordRowProps) {
  const t = useTranslations();
  const c = word.card;
  const ipa = c.ipa?.[accent] || c.ipa?.uk || c.ipa?.us;
  const id = `lib-${word.lexeme_id}`;
  return (
    <li {...rowProps} className={cn("flex items-center gap-3 px-3 py-2.5 sm:px-4", word.in_deck && "bg-muted/40")}>
      {word.in_deck ? (
        <span className="flex size-4 shrink-0 items-center justify-center text-primary" title={t("library.inDeck")}>
          <Check className="size-4" strokeWidth={2} aria-hidden />
        </span>
      ) : (
        <Checkbox id={id} checked={selected} onCheckedChange={onToggle} aria-label={t("library.selectWord", { word: c.headword })} />
      )}
      <Thumb word={word} />
      <label htmlFor={word.in_deck ? undefined : id} className={cn("min-w-0 flex-1", !word.in_deck && "cursor-pointer")}>
        <span className="flex flex-wrap items-baseline gap-x-2 gap-y-0.5">
          <span className="font-medium">{c.headword}</span>
          {ipa && <span className="font-mono text-xs text-muted-foreground">{ipa}</span>}
          {c.part_of_speech && <span className="text-xs italic text-muted-foreground">{c.part_of_speech}</span>}
          {word.cefr && (
            <Badge variant="secondary" className="tabular-nums">
              {word.cefr}
            </Badge>
          )}
          {word.in_deck && <Badge variant="outline">{t("library.inDeck")}</Badge>}
          {!word.in_deck && word.known && <Badge variant="outline">{t("library.known")}</Badge>}
        </span>
        <span className="block truncate text-sm text-primary">{c.meaning_vi}</span>
        {showTopic && <span className="block text-xs text-muted-foreground">{t(`vocab.topics.${word.topic}`)}</span>}
      </label>
      <div className="flex shrink-0 items-center">
        {onSpeak && (
          <Button variant="ghost" size="icon-sm" onClick={onSpeak} aria-label={t("vocab.playWord", { word: c.headword })}>
            <Volume2 className="size-4" strokeWidth={1.75} aria-hidden />
          </Button>
        )}
        <Button variant="ghost" size="icon-sm" onClick={onPreview} aria-label={t("library.viewCard", { word: c.headword })}>
          <Eye className="size-4" strokeWidth={1.75} aria-hidden />
        </Button>
      </div>
    </li>
  );
}

function Thumb({ word }: { word: TLibraryWord }) {
  const img = word.card.image;
  const src = img?.thumb_url || img?.url;
  const [failed, setFailed] = useState(false);
  if (!src || failed) {
    return (
      <span aria-hidden className="flex size-11 shrink-0 items-center justify-center rounded-lg bg-muted text-sm font-semibold text-muted-foreground uppercase">
        {word.card.headword.slice(0, 2)}
      </span>
    );
  }
  return (
    // Card photos are hotlinked from Unsplash or through our Pixabay refresh endpoint.
    // eslint-disable-next-line @next/next/no-img-element
    <img src={src} alt="" loading="lazy" className="size-11 shrink-0 rounded-lg bg-muted object-cover" onError={() => setFailed(true)} />
  );
}
