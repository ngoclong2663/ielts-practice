"use client";

import { Plus } from "lucide-react";
import { useTranslations } from "next-intl";
import { useState } from "react";
import { Button } from "@/components/ui/button";
import { Field, FieldDescription, FieldLabel } from "@/components/ui/field";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Skeleton } from "@/components/ui/skeleton";
import { useIsClient } from "@/hooks/use-is-client";
import { Link } from "@/i18n/navigation";
import { useDecks } from "../hooks/use-decks";
import { CardFormDialog } from "./CardFormDialog";
import { DeckFormDialog } from "./DeckFormDialog";

// Manual card entry: one word at a time, typed by the user.
export function ManualAddPanel() {
  const t = useTranslations();
  const isClient = useIsClient();
  const decksQuery = useDecks();
  const decks = isClient ? decksQuery.data : undefined;
  const [picked, setPicked] = useState<string | null>(null);
  const [adding, setAdding] = useState(false);
  const [creatingDeck, setCreatingDeck] = useState(false);
  const [addedCount, setAddedCount] = useState(0);

  const deckId = picked ?? decks?.[0]?.id ?? null;
  const deckItems = decks?.map((d) => ({ value: d.id, label: d.name })) ?? [];

  return (
    <div>
      <section className="space-y-6 rounded-2xl border border-border bg-card p-6 sm:p-8">
        {!decks ? (
          <Skeleton className="h-9 w-full" />
        ) : decks.length === 0 ? (
          <div className="space-y-3">
            <p className="text-sm text-muted-foreground">{t("addWords.noDecks")}</p>
            <Button onClick={() => setCreatingDeck(true)}>{t("decks.newDeck")}</Button>
          </div>
        ) : (
          <>
            <Field>
              <FieldLabel>{t("addWords.deck")}</FieldLabel>
              <Select items={deckItems} value={deckId} onValueChange={(v) => v && setPicked(String(v))}>
                <SelectTrigger className="w-full sm:w-80" aria-label={t("addWords.deck")}>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  {deckItems.map((d) => (
                    <SelectItem key={d.value} value={d.value}>
                      {d.label}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
              <FieldDescription>
                <button type="button" className="underline underline-offset-2" onClick={() => setCreatingDeck(true)}>
                  {t("addWords.orCreateDeck")}
                </button>
              </FieldDescription>
            </Field>
            <div className="flex flex-wrap items-center gap-3">
              <Button onClick={() => setAdding(true)} disabled={!deckId}>
                <Plus className="size-4" strokeWidth={1.75} aria-hidden />
                {t("cards.addCard")}
              </Button>
              {addedCount > 0 && deckId && (
                <Link href={`/decks/${deckId}`} className="text-sm text-primary underline underline-offset-2">
                  {t("addWords.addedOpenDeck", { count: addedCount })}
                </Link>
              )}
            </div>
          </>
        )}
      </section>

      <CardFormDialog
        open={adding}
        onOpenChange={setAdding}
        deckId={deckId ?? undefined}
        onSaved={() => setAddedCount((n) => n + 1)}
      />
      <DeckFormDialog open={creatingDeck} onOpenChange={setCreatingDeck} onSaved={(deck) => setPicked(deck.id)} />
    </div>
  );
}
