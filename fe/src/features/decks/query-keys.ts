import type { TCardListParams } from "./types";

export const deckKeys = {
  all: ["decks"] as const,
  list: () => [...deckKeys.all, "list"] as const,
  detail: (id: string) => [...deckKeys.all, "detail", id] as const,
  cards: (deckId: string) => [...deckKeys.all, "cards", deckId] as const,
  cardList: (deckId: string, params: TCardListParams) => [...deckKeys.cards(deckId), params] as const,
  card: (id: string) => [...deckKeys.all, "card", id] as const,
};
