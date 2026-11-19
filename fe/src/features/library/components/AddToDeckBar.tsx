"use client";

import { CheckCircle2, GraduationCap, Plus } from "lucide-react";
import { useTranslations } from "next-intl";
import { useState } from "react";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { useDecks } from "@/features/decks/hooks/use-decks";
import { readLastDeck, rememberDeck } from "@/features/vocab/lib/last-deck";
import { useIsClient } from "@/hooks/use-is-client";
import { Link } from "@/i18n/navigation";
import { useAddLibraryWords } from "../hooks/use-library";
import type { TAddLibraryWordsResult } from "../types";

const NEW_DECK = "__new__";

type AddToDeckBarProps = {
  lexemeIds: string[];
  /** Name offered for a new deck (e.g. the topic). */
  defaultDeckName: string;
  onAdded?: (res: TAddLibraryWordsResult) => void;
  onClear?: () => void;
};

/**
 * Sticky footer for ready-made cards (library, writing vocabulary): choose the deck and add the selected words
 * at once, with no AI and no daily limit. After adding, it offers the deck and "Study now".
 */
export function AddToDeckBar({ lexemeIds, defaultDeckName, onAdded, onClear }: AddToDeckBarProps) {
  const t = useTranslations();
  const isClient = useIsClient();
  const decksQuery = useDecks();
  const decks = isClient ? decksQuery.data : undefined;
  const add = useAddLibraryWords();
  const [picked, setPicked] = useState<string | null>(null);
  const [newName, setNewName] = useState<string | null>(null);
  const [added, setAdded] = useState<TAddLibraryWordsResult | null>(null);

  const lastDeck = isClient ? readLastDeck() : null;
  const deckId = picked ?? (lastDeck && decks?.some((d) => d.id === lastDeck) ? lastDeck : NEW_DECK);
  const name = newName ?? defaultDeckName;
  const items = [{ value: NEW_DECK, label: t("vocab.newDeckOption") }, ...(decks ?? []).map((d) => ({ value: d.id, label: d.name }))];

  if (lexemeIds.length === 0 && !added) {
    return null;
  }

  const submit = () =>
    add.mutate(
      { lexeme_ids: lexemeIds, ...(deckId === NEW_DECK ? { new_deck_name: name.trim() } : { deck_id: deckId }) },
      {
        onSuccess: (res) => {
          rememberDeck(res.deck_id);
          setPicked(res.deck_id);
          setAdded(res);
          onAdded?.(res);
          if (res.created > 0) toast.success(t("library.added", { count: res.created, deck: res.deck_name }));
          if (res.skipped.length > 0) {
            toast.info(t("vocab.skipped", { count: res.skipped.length }), { description: res.skipped.join(", ") });
          }
        },
        onError: (error) => toast.error(t(`errors.${error.code}`), { id: "library-add" }),
      },
    );

  return (
    // Sticks above the mobile bottom navigation, and to the bottom of the content column on desktop.
    // A container query, not the viewport: the bar sits in the wide library column and in the narrow attempt column.
    <div className="@container sticky bottom-[calc(4.25rem+env(safe-area-inset-bottom))] z-20 mt-6 lg:bottom-6">
      <div className="flex flex-col gap-3 rounded-2xl border border-border bg-background/95 px-4 py-3 shadow-lg backdrop-blur supports-[backdrop-filter]:bg-background/85 @2xl:flex-row @2xl:items-center @2xl:justify-between">
        {added && lexemeIds.length === 0 ? (
          <>
            <p className="flex items-center gap-2 text-sm">
              <CheckCircle2 className="size-4 shrink-0 text-primary" strokeWidth={1.75} aria-hidden />
              {t("library.addedTo", { count: added.created, deck: added.deck_name })}
            </p>
            <div className="flex gap-2">
              <Button variant="outline" nativeButton={false} render={<Link href={`/decks/${added.deck_id}`} />}>
                {t("vocab.openDeck")}
              </Button>
              <Button nativeButton={false} render={<Link href={`/study?deck=${added.deck_id}`} />}>
                <GraduationCap className="size-4" strokeWidth={1.75} aria-hidden />
                {t("vocab.studyNow")}
              </Button>
            </div>
          </>
        ) : (
          <>
            <div className="flex min-w-0 flex-1 flex-col gap-2 @md:flex-row @md:items-center">
              <span className="shrink-0 text-sm font-medium">{t("vocab.saveTo")}</span>
              <Select items={items} value={deckId} onValueChange={(v) => v && setPicked(String(v))}>
                <SelectTrigger className="w-full @md:w-44 @2xl:w-52" aria-label={t("vocab.saveTo")}>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  {items.map((d) => (
                    <SelectItem key={d.value} value={d.value}>
                      {d.label}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
              {deckId === NEW_DECK && (
                <Input
                  value={name}
                  onChange={(e) => setNewName(e.target.value)}
                  aria-label={t("decks.name")}
                  maxLength={80}
                  className="@md:w-44 @2xl:w-48"
                />
              )}
            </div>
            <div className="flex items-center justify-end gap-2">
              {onClear && (
                <Button variant="ghost" onClick={onClear} disabled={add.isPending}>
                  {t("library.clearSelection")}
                </Button>
              )}
              <Button size="lg" onClick={submit} disabled={add.isPending || (deckId === NEW_DECK && !name.trim())}>
                <Plus className="size-4" strokeWidth={1.75} aria-hidden />
                {add.isPending ? t("common.saving") : t("library.addCount", { count: lexemeIds.length })}
              </Button>
            </div>
          </>
        )}
      </div>
    </div>
  );
}
