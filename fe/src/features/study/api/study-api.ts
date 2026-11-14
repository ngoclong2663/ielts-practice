import { apiClient } from "@/lib/api-client";
import type { TAnswerPayload, TAnswerResponse, TCountsResponse, TStudyQueue, TUndoResponse } from "../types";

export const QUEUE_SIZE = 5;

export function fetchQueue(deckId?: string) {
  const params = new URLSearchParams({ size: String(QUEUE_SIZE) });
  if (deckId) params.set("deck_id", deckId);
  return apiClient.get<TStudyQueue>(`/study/queue?${params.toString()}`);
}

export const answerApi = (p: TAnswerPayload) => apiClient.post<TAnswerResponse>("/study/answer", p);

export const undoApi = (deckId?: string) =>
  apiClient.post<TUndoResponse>("/study/undo", { deck_id: deckId || undefined, size: QUEUE_SIZE });

export const fetchCounts = () => apiClient.get<TCountsResponse>("/study/counts");
