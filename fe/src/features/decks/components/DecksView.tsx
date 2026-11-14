"use client";

import { EllipsisVertical, Layers, Pencil, Plus, Trash2 } from "lucide-react";
import { CardFormDialog } from "./CardFormDialog";
import { useTranslations } from "next-intl";
import { useState } from "react";
import { PageHeader } from "@/components/shared/PageHeader";
import { SrsCounts } from "@/components/srs/SrsCounts";
import { Button, buttonVariants } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Skeleton } from "@/components/ui/skeleton";
import { useStudyCounts } from "@/features/study/hooks/use-study-counts";
import { useIsClient } from "@/hooks/use-is-client";
import { Link } from "@/i18n/navigation";
import { useDecks } from "../hooks/use-decks";
import type { TDeck } from "../types";
import { DeckFormDialog } from "./DeckFormDialog";
import { DeleteDeckDialog } from "./DeleteDeckDialog";

const ZERO = { new: 0, learning: 0, review: 0 };

export function DecksView() {
  const t = useTranslations();
  const isClient = useIsClient();
  const decksQuery = useDecks();
  const countsQuery = useStudyCounts();
  // Until hydration finishes, render exactly what the server rendered (loading state).
  const decks = isClient ? decksQuery : { ...decksQuery, data: undefined, isPending: true, isError: false };
  const counts = isClient ? countsQuery : { data: undefined };
  const [creating, setCreating] = useState(false);
  const [editing, setEditing] = useState<TDeck | null>(null);
  const [deleting, setDeleting] = useState<TDeck | null>(null);
  const [addingTo, setAddingTo] = useState<string | null>(null);

  const countsByDeck = new Map(counts.data?.decks.map((d) => [d.deck_id, d]) ?? []);
  const total = counts.data?.total ?? ZERO;
  const totalDue = total.new + total.learning + total.review;

  return (
    <div className="mx-auto max-w-5xl">
      <PageHeader
        title={t("decks.title")}
        description={t("decks.subtitle")}
        actions={
          <div className="flex gap-2">
            {totalDue > 0 && (
              <Link href="/study" className={buttonVariants({ variant: "outline" })}>
                {t("decks.studyAll")}
              </Link>
            )}
            <Button onClick={() => setCreating(true)}>
              <Plus className="size-4" strokeWidth={1.75} aria-hidden />
              {t("decks.newDeck")}
            </Button>
          </div>
        }
      />

      {decks.isPending && (
        <div className="space-y-3">
          {[0, 1, 2].map((i) => (
            <Skeleton key={i} className="h-20 w-full rounded-xl" />
          ))}
        </div>
      )}

      {decks.isError && (
        <div className="rounded-xl border border-border p-6">
          <p className="text-sm text-muted-foreground">{t("decks.loadError")}</p>
          <Button variant="outline" className="mt-4" onClick={() => decks.refetch()}>
            {t("common.retry")}
          </Button>
        </div>
      )}

      {decks.data?.length === 0 && (
        <div className="flex flex-col items-start gap-4 rounded-2xl border border-dashed border-border p-8 sm:p-10">
          <span className="grid size-11 place-items-center rounded-xl bg-accent text-accent-foreground">
            <Layers className="size-5" strokeWidth={1.75} aria-hidden />
          </span>
          <div>
            <h2 className="text-lg font-semibold tracking-tight">{t("decks.emptyTitle")}</h2>
            <p className="mt-1 text-sm text-muted-foreground">{t("decks.emptyBody")}</p>
          </div>
          <Button onClick={() => setCreating(true)}>{t("decks.newDeck")}</Button>
        </div>
      )}

      {!!decks.data?.length && (
        <ul className="divide-y divide-border overflow-hidden rounded-2xl border border-border bg-card">
          {decks.data.map((deck) => {
            const c = countsByDeck.get(deck.id) ?? ZERO;
            const due = c.new + c.learning + c.review;
            return (
              <li key={deck.id} className="flex items-center gap-3 px-4 py-4 sm:px-5">
                <Link href={`/decks/${deck.id}`} className="min-w-0 flex-1 rounded-md outline-none focus-visible:ring-2 focus-visible:ring-ring">
                  <p className="truncate font-medium">{deck.name}</p>
                  <p className="mt-0.5 text-sm text-muted-foreground">{t("decks.cardCount", { count: deck.card_count })}</p>
                </Link>
                <SrsCounts counts={c} className="hidden sm:inline-flex" />
                <Link
                  href={`/study?deck=${deck.id}`}
                  aria-disabled={due === 0}
                  className={buttonVariants({ variant: due > 0 ? "default" : "outline", size: "sm", className: due === 0 ? "pointer-events-none opacity-50" : "" })}
                >
                  {t("decks.study")}
                  {due > 0 && <span className="font-mono tabular-nums sm:hidden">{due}</span>}
                </Link>
                <DropdownMenu>
                  <DropdownMenuTrigger render={<Button variant="ghost" size="icon-sm" aria-label={t("decks.actions")} />}>
                    <EllipsisVertical className="size-4" strokeWidth={1.75} aria-hidden />
                  </DropdownMenuTrigger>
                  <DropdownMenuContent align="end" className="w-44">
                    <DropdownMenuItem onClick={() => setAddingTo(deck.id)}>
                      <Plus className="size-4" strokeWidth={1.75} aria-hidden /> {t("cards.addCard")}
                    </DropdownMenuItem>
                    <DropdownMenuItem onClick={() => setEditing(deck)}>
                      <Pencil className="size-4" strokeWidth={1.75} aria-hidden /> {t("decks.edit")}
                    </DropdownMenuItem>
                    <DropdownMenuItem variant="destructive" onClick={() => setDeleting(deck)}>
                      <Trash2 className="size-4" strokeWidth={1.75} aria-hidden /> {t("decks.delete")}
                    </DropdownMenuItem>
                  </DropdownMenuContent>
                </DropdownMenu>
              </li>
            );
          })}
        </ul>
      )}

      <DeckFormDialog open={creating} onOpenChange={setCreating} />
      <DeckFormDialog open={!!editing} onOpenChange={(o) => !o && setEditing(null)} deck={editing} />
      <DeleteDeckDialog deck={deleting} onOpenChange={(o) => !o && setDeleting(null)} />
      <CardFormDialog open={!!addingTo} onOpenChange={(o) => !o && setAddingTo(null)} deckId={addingTo ?? undefined} />
    </div>
  );
}
