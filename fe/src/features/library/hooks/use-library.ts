"use client";

import { keepPreviousData, useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { deckKeys } from "@/features/decks/query-keys";
import type { ApiError } from "@/lib/api-client";
import { addLibraryWordsApi, fetchCollections, fetchLibraryWords } from "../api/library-api";
import { libraryKeys } from "../query-keys";
import type { TAddLibraryWordsPayload, TAddLibraryWordsResult, TLibraryWordsParams } from "../types";

// Library content changes only with a release, so it is kept for the whole visit.
const LIBRARY_STALE = 10 * 60 * 1000;

export function useCollections() {
  return useQuery({ queryKey: libraryKeys.collections(), queryFn: fetchCollections, staleTime: LIBRARY_STALE });
}

export function useLibraryWords(slug: string, params: TLibraryWordsParams, enabled = true) {
  return useQuery({
    queryKey: libraryKeys.wordList(slug, params),
    queryFn: () => fetchLibraryWords(slug, params),
    enabled,
    staleTime: LIBRARY_STALE,
    placeholderData: keepPreviousData,
  });
}

/** Adds library cards to a deck; the words' "in deck" flags and the deck lists refresh. */
export function useAddLibraryWords() {
  const qc = useQueryClient();
  return useMutation<TAddLibraryWordsResult, ApiError, TAddLibraryWordsPayload>({
    mutationFn: addLibraryWordsApi,
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: libraryKeys.all });
      void qc.invalidateQueries({ queryKey: deckKeys.all });
    },
  });
}
