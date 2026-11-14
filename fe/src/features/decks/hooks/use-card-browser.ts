import { useTranslations } from "next-intl";
import { useEffect, useState } from "react";
import { toast } from "sonner";
import type { TBulkAction, TCardStateFilter } from "../types";
import { useBulkCards, useCards } from "./use-decks";

/** Search, filter, pagination, selection and bulk actions for one deck's cards. */
export function useCardBrowser(deckId: string) {
  const t = useTranslations("cards");
  const [search, setSearch] = useState("");
  const [q, setQ] = useState("");
  const [state, setState] = useState<TCardStateFilter>("");
  const [page, setPage] = useState(1);
  const [selected, setSelected] = useState<Set<string>>(new Set());
  const bulk = useBulkCards();

  // Debounce typing so every keystroke doesn't hit the API.
  useEffect(() => {
    const id = setTimeout(() => {
      setQ(search.trim());
      setPage(1);
    }, 300);
    return () => clearTimeout(id);
  }, [search]);

  const cards = useCards(deckId, { q, state, page });
  const items = cards.data?.items ?? [];
  const pages = Math.max(1, Math.ceil((cards.data?.total ?? 0) / (cards.data?.page_size ?? 50)));

  const toggle = (id: string) =>
    setSelected((prev) => {
      const next = new Set(prev);
      if (next.has(id)) {
        next.delete(id);
      } else {
        next.add(id);
      }
      return next;
    });

  const allOnPage = items.length > 0 && items.every((c) => selected.has(c.id));
  const toggleAll = () => setSelected(allOnPage ? new Set() : new Set(items.map((c) => c.id)));

  const run = (action: TBulkAction, targetDeckId?: string) =>
    bulk.mutate(
      { card_ids: [...selected], action, deck_id: targetDeckId },
      {
        onSuccess: ({ affected }) => {
          toast.success(t("bulkDone", { count: affected }));
          setSelected(new Set());
        },
      },
    );

  return {
    search,
    setSearch,
    state,
    setState: (s: TCardStateFilter) => {
      setState(s);
      setPage(1);
      setSelected(new Set());
    },
    page,
    pages,
    setPage,
    cards,
    items,
    selected,
    toggle,
    allOnPage,
    toggleAll,
    run,
    isBulkPending: bulk.isPending,
    hasFilters: !!q || !!state,
  };
}
