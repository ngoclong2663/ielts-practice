import { keepPreviousData, useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import type { ApiError } from "@/lib/api-client";
import { studyKeys } from "@/features/study/query-keys";
import {
  bulkCardsApi,
  createCardApi,
  createDeckApi,
  deleteDeckApi,
  fetchCard,
  fetchCards,
  fetchDeck,
  fetchDecks,
  updateCardApi,
  updateDeckApi,
} from "../api/decks-api";
import { deckKeys } from "../query-keys";
import type {
  TBulkAction,
  TCard,
  TCardListParams,
  TCreateDeckPayload,
  TDeck,
  TManualCardFields,
  TUpdateDeckPayload,
} from "../types";

export function useDecks() {
  return useQuery({ queryKey: deckKeys.list(), queryFn: fetchDecks });
}

export function useDeck(id: string) {
  return useQuery({ queryKey: deckKeys.detail(id), queryFn: () => fetchDeck(id), enabled: !!id });
}

export function useCards(deckId: string, params: TCardListParams) {
  return useQuery({
    queryKey: deckKeys.cardList(deckId, params),
    queryFn: () => fetchCards(deckId, params),
    placeholderData: keepPreviousData,
  });
}

export function useCard(id: string | null) {
  return useQuery({ queryKey: deckKeys.card(id ?? ""), queryFn: () => fetchCard(id as string), enabled: !!id });
}

// Anything that changes cards also changes due counts.
function useInvalidateStudyData() {
  const qc = useQueryClient();
  return () => {
    qc.invalidateQueries({ queryKey: deckKeys.all });
    qc.invalidateQueries({ queryKey: studyKeys.all });
  };
}

export function useCreateDeck() {
  const invalidate = useInvalidateStudyData();
  return useMutation<TDeck, ApiError, TCreateDeckPayload>({ mutationFn: createDeckApi, onSuccess: invalidate });
}

export function useUpdateDeck(id: string) {
  const invalidate = useInvalidateStudyData();
  return useMutation<TDeck, ApiError, TUpdateDeckPayload>({
    mutationFn: (p) => updateDeckApi(id, p),
    onSuccess: invalidate,
  });
}

export function useDeleteDeck() {
  const invalidate = useInvalidateStudyData();
  return useMutation<void, ApiError, string>({ mutationFn: deleteDeckApi, onSuccess: invalidate });
}

export function useCreateCard(deckId: string) {
  const invalidate = useInvalidateStudyData();
  return useMutation<TCard, ApiError, TManualCardFields & { allow_duplicate?: boolean }>({
    mutationFn: (p) => createCardApi(deckId, p),
    onSuccess: invalidate,
  });
}

export function useUpdateCard() {
  const qc = useQueryClient();
  const invalidate = useInvalidateStudyData();
  return useMutation<TCard, ApiError, { id: string; fields: TManualCardFields }>({
    mutationFn: ({ id, fields }) => updateCardApi(id, fields),
    onSuccess: (card) => {
      qc.setQueryData(deckKeys.card(card.id), card);
      invalidate();
    },
  });
}

export function useBulkCards() {
  const invalidate = useInvalidateStudyData();
  return useMutation<{ affected: number }, ApiError, { card_ids: string[]; action: TBulkAction; deck_id?: string }>({
    mutationFn: bulkCardsApi,
    onSuccess: invalidate,
  });
}
