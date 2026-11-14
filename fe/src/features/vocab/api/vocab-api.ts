import { apiClient } from "@/lib/api-client";
import type {
  TCommitPayload,
  TCommitResult,
  TEntry,
  TImageSearchResult,
  TJob,
  TJobItem,
  TJobSummary,
  TReportReason,
  TRetryPayload,
  TUpdateItemPayload,
  TVocabStatus,
} from "../types";

export const fetchVocabStatus = () => apiClient.get<TVocabStatus>("/vocab/status");
export const fetchJobs = () => apiClient.get<{ items: TJobSummary[] }>("/vocab/jobs").then((r) => r.items);
export const fetchJob = (id: string) => apiClient.get<TJob>(`/vocab/jobs/${id}`);

export const createJobApi = (p: { entries: TEntry[]; target_deck_id?: string }) => apiClient.post<TJob>("/vocab/jobs", p);

const itemPath = (jobId: string, itemId: string) => `/vocab/jobs/${jobId}/items/${itemId}`;

export const retryItemApi = (jobId: string, itemId: string, p: TRetryPayload) =>
  apiClient.post<TJobItem>(`${itemPath(jobId, itemId)}/retry`, p);

export const updateItemApi = (jobId: string, itemId: string, p: TUpdateItemPayload) =>
  apiClient.patch<TJobItem>(itemPath(jobId, itemId), p);

export function searchImagesApi(jobId: string, itemId: string, q: string, provider?: string) {
  const params = new URLSearchParams();
  if (q) params.set("q", q);
  if (provider) params.set("provider", provider);
  return apiClient.get<TImageSearchResult>(`${itemPath(jobId, itemId)}/images?${params.toString()}`);
}

export const commitJobApi = (jobId: string, p: TCommitPayload) =>
  apiClient.post<TCommitResult>(`/vocab/jobs/${jobId}/commit`, p);

export const reportLexemeApi = (lexemeId: string, p: { reason: TReportReason; note?: string }) =>
  apiClient.post<void>(`/lexemes/${lexemeId}/report`, p);
