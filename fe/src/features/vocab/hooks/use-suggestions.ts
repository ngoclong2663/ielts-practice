import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import type { ApiError } from "@/lib/api-client";
import {
  createSuggestionApi,
  fetchSuggestion,
  fetchTopics,
  markKnownApi,
  moreSuggestionsApi,
  unmarkKnownApi,
} from "../api/vocab-api";
import { vocabKeys } from "../query-keys";
import type { TCreateSuggestionPayload, TSuggestion } from "../types";

export function useTopics() {
  return useQuery({ queryKey: vocabKeys.topics(), queryFn: fetchTopics, staleTime: Infinity });
}

export function useSuggestion(id: string | null) {
  return useQuery({ queryKey: vocabKeys.suggestion(id ?? ""), queryFn: () => fetchSuggestion(id as string), enabled: !!id });
}

export function useCreateSuggestion() {
  const qc = useQueryClient();
  return useMutation<TSuggestion, ApiError, TCreateSuggestionPayload>({
    mutationFn: createSuggestionApi,
    onSuccess: (s) => {
      qc.setQueryData(vocabKeys.suggestion(s.id), s);
      qc.invalidateQueries({ queryKey: vocabKeys.status() });
    },
  });
}

export function useMoreSuggestions(id: string) {
  const qc = useQueryClient();
  return useMutation<TSuggestion, ApiError, number>({
    mutationFn: (count) => moreSuggestionsApi(id, count),
    onSuccess: (s) => {
      qc.setQueryData(vocabKeys.suggestion(id), s);
      qc.invalidateQueries({ queryKey: vocabKeys.status() });
    },
  });
}

/** "I know this": hides the word from this list now and from every later suggestion. */
export function useKnownWord(suggestionId: string) {
  const qc = useQueryClient();
  const key = vocabKeys.suggestion(suggestionId);
  const mark = useMutation<void, ApiError, string>({
    mutationFn: markKnownApi,
    onMutate: (lemma) =>
      qc.setQueryData<TSuggestion>(key, (s) => (s ? { ...s, words: s.words.filter((w) => w.lemma !== lemma) } : s)),
  });
  const unmark = useMutation<void, ApiError, string>({
    mutationFn: unmarkKnownApi,
    onSuccess: () => qc.invalidateQueries({ queryKey: key }),
  });
  return { mark, unmark };
}
