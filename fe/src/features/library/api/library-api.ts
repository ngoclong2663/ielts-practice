import { apiClient } from "@/lib/api-client";
import type {
  TAddLibraryWordsPayload,
  TAddLibraryWordsResult,
  TLibraryCollection,
  TLibraryWords,
  TLibraryWordsParams,
} from "../types";

export const fetchCollections = () =>
  apiClient.get<{ collections: TLibraryCollection[] }>("/library/collections").then((r) => r.collections);

export function fetchLibraryWords(slug: string, params: TLibraryWordsParams) {
  const q = new URLSearchParams();
  if (params.topic) q.set("topic", params.topic);
  if (params.q) q.set("q", params.q);
  return apiClient.get<TLibraryWords>(`/library/collections/${encodeURIComponent(slug)}/words?${q.toString()}`);
}

export const addLibraryWordsApi = (p: TAddLibraryWordsPayload) => apiClient.post<TAddLibraryWordsResult>("/library/add", p);
