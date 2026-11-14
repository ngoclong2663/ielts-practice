"use client";

import { ChevronLeft, ChevronRight, Search } from "lucide-react";
import { useFormatter, useTranslations } from "next-intl";
import { useState } from "react";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Checkbox } from "@/components/ui/checkbox";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Input } from "@/components/ui/input";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Skeleton } from "@/components/ui/skeleton";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@/components/ui/alert-dialog";
import { useIsClient } from "@/hooks/use-is-client";
import { useCardBrowser } from "../hooks/use-card-browser";
import { useDecks } from "../hooks/use-decks";
import type { TCardStateFilter, TCardSummary } from "../types";

type CardBrowserProps = {
  deckId: string;
  onEdit: (cardId: string) => void;
};

const STATE_FILTERS: TCardStateFilter[] = ["", "new", "learning", "review", "suspended", "leech"];

function StateBadge({ card }: { card: TCardSummary }) {
  const t = useTranslations("srs");
  if (card.suspended) return <Badge variant="outline">{t("suspended")}</Badge>;
  const color =
    card.state === "new" ? "text-srs-new" : card.state === "review" ? "text-srs-review" : "text-srs-learn";
  return (
    <span className="inline-flex items-center gap-2">
      <span className={`text-sm font-medium ${color}`}>{t(card.state)}</span>
      {card.is_leech && <Badge variant="secondary">{t("leech")}</Badge>}
    </span>
  );
}

export function CardBrowser({ deckId, onEdit }: CardBrowserProps) {
  const t = useTranslations();
  const format = useFormatter();
  const b = useCardBrowser(deckId);
  const isClient = useIsClient();
  const decks = useDecks();
  const [confirmDelete, setConfirmDelete] = useState(false);
  const otherDecks = decks.data?.filter((d) => d.id !== deckId) ?? [];

  const stateItems = STATE_FILTERS.map((s) => ({ value: s, label: s ? t(`srs.${s}`) : t("cards.allStates") }));

  return (
    <section className="space-y-4">
      <div className="flex flex-col gap-3 sm:flex-row">
        <div className="relative flex-1">
          <Search className="pointer-events-none absolute top-1/2 left-3 size-4 -translate-y-1/2 text-muted-foreground" aria-hidden />
          <Input
            value={b.search}
            onChange={(e) => b.setSearch(e.target.value)}
            placeholder={t("cards.searchPlaceholder")}
            aria-label={t("cards.searchPlaceholder")}
            className="pl-9"
          />
        </div>
        <Select items={stateItems} value={b.state} onValueChange={(v) => b.setState((v ?? "") as TCardStateFilter)}>
          <SelectTrigger className="w-full sm:w-48" aria-label={t("cards.state")}>
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            {stateItems.map((s) => (
              <SelectItem key={s.value || "all"} value={s.value}>
                {s.label}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
      </div>

      {b.selected.size > 0 && (
        <div className="flex flex-wrap items-center gap-2 rounded-xl bg-secondary px-3 py-2">
          <span className="mr-2 text-sm font-medium">{t("cards.selected", { count: b.selected.size })}</span>
          <Button size="sm" variant="outline" disabled={b.isBulkPending} onClick={() => b.run("suspend")}>
            {t("cards.suspend")}
          </Button>
          <Button size="sm" variant="outline" disabled={b.isBulkPending} onClick={() => b.run("unsuspend")}>
            {t("cards.unsuspend")}
          </Button>
          <Button size="sm" variant="outline" disabled={b.isBulkPending} onClick={() => b.run("reset")}>
            {t("cards.reset")}
          </Button>
          {otherDecks.length > 0 && (
            <DropdownMenu>
              <DropdownMenuTrigger render={<Button size="sm" variant="outline" disabled={b.isBulkPending} />}>
                {t("cards.moveTo")}
              </DropdownMenuTrigger>
              <DropdownMenuContent className="w-56">
                {otherDecks.map((d) => (
                  <DropdownMenuItem key={d.id} onClick={() => b.run("move", d.id)}>
                    <span className="truncate">{d.name}</span>
                  </DropdownMenuItem>
                ))}
              </DropdownMenuContent>
            </DropdownMenu>
          )}
          <Button size="sm" variant="destructive" disabled={b.isBulkPending} onClick={() => setConfirmDelete(true)}>
            {t("cards.delete")}
          </Button>
        </div>
      )}

      {!isClient || b.cards.isPending ? (
        <div className="space-y-2">
          {[0, 1, 2, 3].map((i) => (
            <Skeleton key={i} className="h-11 w-full" />
          ))}
        </div>
      ) : b.items.length === 0 ? (
        <p className="rounded-xl border border-dashed border-border p-8 text-center text-sm text-muted-foreground">
          {b.hasFilters ? t("cards.noResults") : t("cards.emptyBody")}
        </p>
      ) : (
        <div className="overflow-hidden rounded-xl border border-border">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead className="w-10">
                  <Checkbox checked={b.allOnPage} onCheckedChange={b.toggleAll} aria-label={t("cards.selectAll")} />
                </TableHead>
                <TableHead>{t("cards.word")}</TableHead>
                <TableHead className="hidden md:table-cell">{t("cards.meaning")}</TableHead>
                <TableHead>{t("cards.state")}</TableHead>
                <TableHead className="hidden sm:table-cell">{t("cards.due")}</TableHead>
                <TableHead className="hidden text-right lg:table-cell">{t("cards.lapses")}</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {b.items.map((card) => (
                <TableRow key={card.id} data-state={b.selected.has(card.id) ? "selected" : undefined}>
                  <TableCell>
                    <Checkbox
                      checked={b.selected.has(card.id)}
                      onCheckedChange={() => b.toggle(card.id)}
                      aria-label={t("cards.selectCard", { word: card.headword })}
                    />
                  </TableCell>
                  <TableCell>
                    <button type="button" onClick={() => onEdit(card.id)} className="text-left font-medium hover:text-primary hover:underline">
                      {card.headword}
                    </button>
                    <p className="max-w-[16rem] truncate text-sm text-muted-foreground md:hidden">{card.meaning_vi}</p>
                  </TableCell>
                  <TableCell className="hidden max-w-xs truncate text-muted-foreground md:table-cell">{card.meaning_vi}</TableCell>
                  <TableCell>
                    <StateBadge card={card} />
                  </TableCell>
                  <TableCell className="hidden text-sm text-muted-foreground sm:table-cell">
                    {card.state === "new" ? "" : format.relativeTime(new Date(card.due))}
                  </TableCell>
                  <TableCell className="hidden text-right font-mono tabular-nums lg:table-cell">{card.lapses}</TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </div>
      )}

      {b.pages > 1 && (
        <div className="flex items-center justify-end gap-2">
          <span className="mr-2 text-sm text-muted-foreground">{t("cards.page", { page: b.page, pages: b.pages })}</span>
          <Button size="icon-sm" variant="outline" disabled={b.page <= 1} onClick={() => b.setPage(b.page - 1)} aria-label={t("cards.previous")}>
            <ChevronLeft className="size-4" aria-hidden />
          </Button>
          <Button size="icon-sm" variant="outline" disabled={b.page >= b.pages} onClick={() => b.setPage(b.page + 1)} aria-label={t("cards.next")}>
            <ChevronRight className="size-4" aria-hidden />
          </Button>
        </div>
      )}

      <AlertDialog open={confirmDelete} onOpenChange={setConfirmDelete}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>{t("cards.deleteConfirmTitle", { count: b.selected.size })}</AlertDialogTitle>
            <AlertDialogDescription>{t("cards.deleteConfirmBody")}</AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>{t("common.cancel")}</AlertDialogCancel>
            <AlertDialogAction
              variant="destructive"
              onClick={() => {
                b.run("delete");
                setConfirmDelete(false);
              }}
            >
              {t("cards.delete")}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </section>
  );
}
