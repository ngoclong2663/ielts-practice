import { apiClient } from "@/lib/api-client";
import type {
  TBulkAction,
  TCard,
  TCardList,
  TCardListParams,
  TCreateDeckPayload,
  TDeck,
  TManualCardFields,
  TUpdateDeckPayload,
} from "../types";

export const fetchDecks = () => apiClient.get<{ items: TDeck[] }>("/decks").then((r) => r.items);
export const fetchDeck = (id: string) => apiClient.get<TDeck>(`/decks/${id}`);
export const createDeckApi = (p: TCreateDeckPayload) => apiClient.post<TDeck>("/decks", p);
export const updateDeckApi = (id: string, p: TUpdateDeckPayload) => apiClient.patch<TDeck>(`/decks/${id}`, p);
export const deleteDeckApi = (id: string) => apiClient.delete<void>(`/decks/${id}`);

export function fetchCards(deckId: string, { q, state, page }: TCardListParams) {
  const params = new URLSearchParams({ page: String(page), page_size: "50" });
  if (q) params.set("q", q);
  if (state) params.set("state", state);
  return apiClient.get<TCardList>(`/decks/${deckId}/cards?${params.toString()}`);
}

export const fetchCard = (id: string) => apiClient.get<TCard>(`/cards/${id}`);
export const createCardApi = (deckId: string, p: TManualCardFields & { allow_duplicate?: boolean }) =>
  apiClient.post<TCard>(`/decks/${deckId}/cards`, p);
export const updateCardApi = (id: string, p: TManualCardFields) => apiClient.patch<TCard>(`/cards/${id}`, p);
export const bulkCardsApi = (p: { card_ids: string[]; action: TBulkAction; deck_id?: string }) =>
  apiClient.post<{ affected: number }>("/cards/bulk", p);
