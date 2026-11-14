"use client";

import { ArrowLeft, Pencil, Plus, Trash2 } from "lucide-react";
import { useTranslations } from "next-intl";
import { useParams } from "next/navigation";
import { useState } from "react";
import { SrsCounts } from "@/components/srs/SrsCounts";
import { Button, buttonVariants } from "@/components/ui/button";
import { Skeleton } from "@/components/ui/skeleton";
import { useStudyCounts } from "@/features/study/hooks/use-study-counts";
import { useIsClient } from "@/hooks/use-is-client";
import { Link, useRouter } from "@/i18n/navigation";
import { useDeck } from "../hooks/use-decks";
import { CardBrowser } from "./CardBrowser";
import { CardFormDialog } from "./CardFormDialog";
import { DeckFormDialog } from "./DeckFormDialog";
import { DeleteDeckDialog } from "./DeleteDeckDialog";

export function DeckDetailView() {
  const t = useTranslations();
  const router = useRouter();
  const { deckId } = useParams<{ deckId: string }>();
  const isClient = useIsClient();
  const deckQuery = useDeck(deckId);
  const countsQuery = useStudyCounts();
  // Until hydration finishes, render the loading state the server rendered.
  const deck = isClient ? deckQuery : { data: undefined, isError: false };
  const counts = isClient ? countsQuery : { data: undefined };
  const [adding, setAdding] = useState(false);
  const [editingCard, setEditingCard] = useState<string | null>(null);
  const [editingDeck, setEditingDeck] = useState(false);
  const [deleting, setDeleting] = useState(false);

  const c = counts.data?.decks.find((d) => d.deck_id === deckId) ?? { new: 0, learning: 0, review: 0 };
  const due = c.new + c.learning + c.review;

  if (deck.isError) {
    return (
      <div className="mx-auto max-w-5xl">
        <p className="text-muted-foreground">{t("errors.DECK_NOT_FOUND")}</p>
        <Link href="/decks" className={buttonVariants({ variant: "outline", className: "mt-4" })}>
          {t("decks.back")}
        </Link>
      </div>
    );
  }

  return (
    <div className="mx-auto max-w-5xl">
      <Link href="/decks" className="mb-4 inline-flex items-center gap-1.5 text-sm text-muted-foreground hover:text-foreground">
        <ArrowLeft className="size-4" aria-hidden /> {t("decks.back")}
      </Link>

      <div className="mb-8 flex flex-col gap-4 sm:flex-row sm:items-end sm:justify-between">
        <div className="min-w-0">
          {deck.data ? (
            <h1 className="truncate text-2xl font-semibold tracking-tight sm:text-3xl">{deck.data.name}</h1>
          ) : (
            <Skeleton className="h-9 w-64" />
          )}
          <div className="mt-2 flex items-center gap-4 text-sm text-muted-foreground">
            {deck.data && <span>{t("decks.cardCount", { count: deck.data.card_count })}</span>}
            <SrsCounts counts={c} />
          </div>
        </div>
        <div className="flex flex-wrap gap-2">
          <Button variant="ghost" size="icon" onClick={() => setEditingDeck(true)} aria-label={t("decks.edit")} disabled={!deck.data}>
            <Pencil className="size-4" strokeWidth={1.75} aria-hidden />
          </Button>
          <Button variant="ghost" size="icon" onClick={() => setDeleting(true)} aria-label={t("decks.delete")} disabled={!deck.data}>
            <Trash2 className="size-4" strokeWidth={1.75} aria-hidden />
          </Button>
          <Button variant="outline" onClick={() => setAdding(true)}>
            <Plus className="size-4" strokeWidth={1.75} aria-hidden /> {t("cards.addCard")}
          </Button>
          <Link
            href={`/study?deck=${deckId}`}
            aria-disabled={due === 0}
            className={buttonVariants({ className: due === 0 ? "pointer-events-none opacity-50" : "" })}
          >
            {t("decks.study")}
          </Link>
        </div>
      </div>

      <CardBrowser deckId={deckId} onEdit={setEditingCard} onAdd={() => setAdding(true)} />

      <CardFormDialog open={adding} onOpenChange={setAdding} deckId={deckId} />
      <CardFormDialog open={!!editingCard} onOpenChange={(o) => !o && setEditingCard(null)} cardId={editingCard} />
      <DeckFormDialog open={editingDeck} onOpenChange={setEditingDeck} deck={deck.data ?? null} />
      <DeleteDeckDialog
        deck={deleting ? (deck.data ?? null) : null}
        onOpenChange={(o) => !o && setDeleting(false)}
        onDeleted={() => router.replace("/decks")}
      />
    </div>
  );
}
