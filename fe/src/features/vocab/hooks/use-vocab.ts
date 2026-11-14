import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { deckKeys } from "@/features/decks/query-keys";
import { studyKeys } from "@/features/study/query-keys";
import type { ApiError } from "@/lib/api-client";
import {
  commitJobApi,
  createJobApi,
  fetchJob,
  fetchJobs,
  fetchVocabStatus,
  reportLexemeApi,
  retryItemApi,
  searchImagesApi,
  updateItemApi,
} from "../api/vocab-api";
import { vocabKeys } from "../query-keys";
import {
  FINAL_ITEM_STATUSES,
  type TCommitPayload,
  type TCommitResult,
  type TEntry,
  type TImageSearchResult,
  type TJob,
  type TJobItem,
  type TReportReason,
  type TRetryPayload,
  type TUpdateItemPayload,
} from "../types";

const POLL_MS = 1500;

export function useVocabStatus() {
  return useQuery({ queryKey: vocabKeys.status(), queryFn: fetchVocabStatus });
}

export function useRecentJobs() {
  return useQuery({ queryKey: vocabKeys.jobs(), queryFn: fetchJobs });
}

const isWorking = (job?: TJob) => !!job?.items.some((i) => !i.removed && !FINAL_ITEM_STATUSES.includes(i.status));

/** Polls the job every 1.5 s while any word is still being generated (Epic 3 §5.2). */
export function useJob(jobId: string | null) {
  return useQuery({
    queryKey: vocabKeys.job(jobId ?? ""),
    queryFn: () => fetchJob(jobId as string),
    enabled: !!jobId,
    refetchInterval: (query) => (isWorking(query.state.data) ? POLL_MS : false),
  });
}

/** Replaces one item inside the cached job so edits show without waiting for the next poll. */
function usePatchJobItem(jobId: string) {
  const qc = useQueryClient();
  return (item: TJobItem) =>
    qc.setQueryData<TJob>(vocabKeys.job(jobId), (job) =>
      job ? { ...job, items: job.items.map((i) => (i.id === item.id ? item : i)) } : job,
    );
}

export function useCreateJob() {
  const qc = useQueryClient();
  return useMutation<TJob, ApiError, { entries: TEntry[]; target_deck_id?: string }>({
    mutationFn: createJobApi,
    onSuccess: (job) => {
      qc.setQueryData(vocabKeys.job(job.id), job);
      qc.invalidateQueries({ queryKey: vocabKeys.jobs() });
      qc.invalidateQueries({ queryKey: vocabKeys.status() });
    },
  });
}

export function useRetryItem(jobId: string) {
  const qc = useQueryClient();
  const patch = usePatchJobItem(jobId);
  return useMutation<TJobItem, ApiError, { itemId: string } & TRetryPayload>({
    mutationFn: ({ itemId, ...p }) => retryItemApi(jobId, itemId, p),
    onSuccess: (item) => {
      patch(item);
      qc.invalidateQueries({ queryKey: vocabKeys.job(jobId) });
      qc.invalidateQueries({ queryKey: vocabKeys.status() });
    },
  });
}

export function useUpdateItem(jobId: string) {
  const qc = useQueryClient();
  const patch = usePatchJobItem(jobId);
  return useMutation<TJobItem, ApiError, { itemId: string } & TUpdateItemPayload>({
    mutationFn: ({ itemId, ...p }) => updateItemApi(jobId, itemId, p),
    onSuccess: (item, vars) => {
      patch(item);
      if (vars.removed !== undefined) {
        qc.invalidateQueries({ queryKey: vocabKeys.job(jobId) });
      }
    },
  });
}

export function useSearchImages(jobId: string) {
  return useMutation<TImageSearchResult, ApiError, { itemId: string; q: string; provider?: string }>({
    mutationFn: ({ itemId, q, provider }) => searchImagesApi(jobId, itemId, q, provider),
  });
}

export function useCommitJob(jobId: string) {
  const qc = useQueryClient();
  return useMutation<TCommitResult, ApiError, TCommitPayload>({
    mutationFn: (p) => commitJobApi(jobId, p),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: vocabKeys.all });
      qc.invalidateQueries({ queryKey: deckKeys.all });
      qc.invalidateQueries({ queryKey: studyKeys.all });
    },
  });
}

export function useReportLexeme() {
  return useMutation<void, ApiError, { lexemeId: string; reason: TReportReason; note?: string }>({
    mutationFn: ({ lexemeId, ...p }) => reportLexemeApi(lexemeId, p),
  });
}
