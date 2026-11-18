import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import type { ApiError } from "@/lib/api-client";
import {
  answerPlanApi,
  checkParagraphApi,
  fetchAttempt,
  fetchAttemptVocab,
  fetchAttempts,
  fetchTask,
  fetchTasks,
  fetchWritingStatus,
  startAttemptApi,
  submitAttemptApi,
} from "../api/writing-api";
import { writingKeys } from "../query-keys";
import type { TAttempt, TLevel, TPosition } from "../types";

export function useWritingStatus() {
  return useQuery({ queryKey: writingKeys.status(), queryFn: fetchWritingStatus });
}

export function useTasks(taskNo: number) {
  return useQuery({ queryKey: writingKeys.tasks(taskNo), queryFn: () => fetchTasks(taskNo) });
}

export function useTask(id: string) {
  return useQuery({ queryKey: writingKeys.task(id), queryFn: () => fetchTask(id), enabled: !!id });
}

export function useAttempts() {
  return useQuery({ queryKey: writingKeys.attempts(), queryFn: fetchAttempts });
}

export function useAttempt(id: string) {
  return useQuery({ queryKey: writingKeys.attempt(id), queryFn: () => fetchAttempt(id), enabled: !!id });
}

/** The lesson's useful vocabulary. The first request for a lesson writes it (an AI call), so it never refetches. */
export function useAttemptVocab(attemptId: string, enabled: boolean) {
  return useQuery({
    queryKey: writingKeys.vocab(attemptId),
    queryFn: () => fetchAttemptVocab(attemptId),
    enabled: enabled && !!attemptId,
    staleTime: Infinity,
    retry: 1,
  });
}

/** Every attempt mutation returns the whole attempt, which replaces the cached one. */
function useAttemptMutation<TVars>(id: string, fn: (vars: TVars) => Promise<TAttempt>, affectsLimits = false) {
  const qc = useQueryClient();
  return useMutation<TAttempt, ApiError, TVars>({
    mutationFn: fn,
    onSuccess: (attempt) => {
      qc.setQueryData(writingKeys.attempt(id), attempt);
      if (affectsLimits) qc.invalidateQueries({ queryKey: writingKeys.status() });
    },
  });
}

export function useStartAttempt(taskId: string) {
  const qc = useQueryClient();
  return useMutation<TAttempt, ApiError, { level: TLevel; position?: TPosition }>({
    mutationFn: (p) => startAttemptApi(taskId, p),
    onSuccess: (attempt) => {
      qc.setQueryData(writingKeys.attempt(attempt.id), attempt);
      qc.invalidateQueries({ queryKey: writingKeys.all });
    },
  });
}

export const useAnswerPlan = (id: string) => useAttemptMutation<number[]>(id, (answers) => answerPlanApi(id, answers));

export const useCheckParagraph = (id: string) =>
  useAttemptMutation<{ key: string; answers: string[] }>(id, (v) => checkParagraphApi(id, v.key, v.answers), true);

export function useSubmitAttempt(id: string) {
  const qc = useQueryClient();
  return useMutation<TAttempt, ApiError, { text: string; seconds_spent: number }>({
    mutationFn: (p) => submitAttemptApi(id, p),
    onSuccess: (attempt) => {
      qc.setQueryData(writingKeys.attempt(id), attempt);
      qc.invalidateQueries({ queryKey: writingKeys.all, predicate: (q) => q.queryKey[1] !== "attempt" });
    },
  });
}
