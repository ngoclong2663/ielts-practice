"use client";

import { useTranslations } from "next-intl";
import { useState } from "react";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { DeckFormDialog } from "@/features/decks/components/DeckFormDialog";
import { useDecks } from "@/features/decks/hooks/use-decks";
import { useIsClient } from "@/hooks/use-is-client";
import { readLastDeck } from "../lib/last-deck";

/** "New deck" means the deck is named later, in the save bar. */
export const NEW_DECK = "__new__";

/** The target deck while adding words: the picked deck, else the last deck used, else a new deck. */
export function useTargetDeck() {
  const isClient = useIsClient();
  const decksQuery = useDecks();
  const decks = isClient ? decksQuery.data : undefined;
  const [picked, setPicked] = useState<string | null>(null);
  const lastDeck = isClient ? readLastDeck() : null;
  const deckId = picked ?? (lastDeck && decks?.some((d) => d.id === lastDeck) ? lastDeck : NEW_DECK);
  return { decks, deckId, setDeckId: setPicked, targetDeckId: deckId === NEW_DECK ? undefined : deckId };
}

type DeckPickerProps = {
  id: string;
  target: ReturnType<typeof useTargetDeck>;
};

/** Deck select plus "or create a new deck", shared by the paste and topic tabs. */
export function DeckPicker({ id, target }: DeckPickerProps) {
  const t = useTranslations();
  const [creating, setCreating] = useState(false);
  const items = [
    { value: NEW_DECK, label: t("vocab.newDeckOption") },
    ...(target.decks ?? []).map((d) => ({ value: d.id, label: d.name })),
  ];
  return (
    <div className="flex flex-col gap-2 sm:flex-row sm:items-center">
      <label htmlFor={id} className="text-sm font-medium">
        {t("vocab.saveTo")}
      </label>
      <Select items={items} value={target.deckId} onValueChange={(v) => v && target.setDeckId(String(v))}>
        <SelectTrigger id={id} className="w-full sm:w-64" aria-label={t("vocab.saveTo")}>
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
      <button
        type="button"
        className="text-left text-sm text-muted-foreground underline underline-offset-2 hover:text-foreground"
        onClick={() => setCreating(true)}
      >
        {t("addWords.orCreateDeck")}
      </button>
      <DeckFormDialog open={creating} onOpenChange={setCreating} onSaved={(deck) => target.setDeckId(deck.id)} />
    </div>
  );
}
