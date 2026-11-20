import { useInfiniteQuery, useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import type { ApiError } from "@/lib/api-client";
import { fetchAttempt, fetchAttempts, fetchBooks, fetchStats, fetchTest, submitAttemptApi } from "../api/exam-api";
import { examKeys } from "../query-keys";
import type { TSkill, TSubmitAttempt } from "../types";

export function useBooks(skill: TSkill) {
  return useQuery({ queryKey: examKeys.books(skill), queryFn: () => fetchBooks(skill) });
}

/** Test content never changes during a run, so it is not refetched in the background. */
export function useTest(slug: string, skill: TSkill) {
  return useQuery({
    queryKey: examKeys.test(slug, skill),
    queryFn: () => fetchTest(slug, skill),
    enabled: !!slug,
    staleTime: Infinity,
  });
}

export function useAttempt(id: string) {
  return useQuery({ queryKey: examKeys.attempt(id), queryFn: () => fetchAttempt(id), enabled: !!id, staleTime: Infinity });
}

export function useAttempts(skill: TSkill) {
  return useInfiniteQuery({
    queryKey: examKeys.attempts(skill),
    queryFn: ({ pageParam }) => fetchAttempts(skill, pageParam),
    initialPageParam: undefined as string | undefined,
    getNextPageParam: (last) => last.next_before,
  });
}

export function useStats(skill: TSkill) {
  return useQuery({ queryKey: examKeys.stats(skill), queryFn: () => fetchStats(skill) });
}

export function useSubmitAttempt() {
  const qc = useQueryClient();
  return useMutation<{ id: string }, ApiError, TSubmitAttempt>({
    mutationFn: submitAttemptApi,
    onSuccess: (_, vars) => {
      qc.invalidateQueries({ queryKey: examKeys.books(vars.skill) });
      qc.invalidateQueries({ queryKey: examKeys.attempts(vars.skill) });
      qc.invalidateQueries({ queryKey: examKeys.stats(vars.skill) });
    },
  });
}
