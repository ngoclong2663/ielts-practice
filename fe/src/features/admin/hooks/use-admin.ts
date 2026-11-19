import { keepPreviousData, useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import type { ApiError } from "@/lib/api-client";
import {
  approveCandidatesApi,
  fetchAdminCandidates,
  fetchAdminFeedback,
  fetchAdminOverview,
  fetchAdminUsers,
  rejectCandidatesApi,
  updateFeedbackApi,
} from "../api/admin-api";
import { adminKeys } from "../query-keys";
import type { TAdminFeedback, TAdminFeedbackQuery, TAdminUsersQuery, TApproveCandidatesRequest, TUpdateFeedbackRequest } from "../types";

export function useAdminOverview() {
  return useQuery({ queryKey: adminKeys.overview(), queryFn: fetchAdminOverview, staleTime: 30 * 1000 });
}

export function useAdminUsers(q: TAdminUsersQuery) {
  return useQuery({ queryKey: adminKeys.users(q), queryFn: () => fetchAdminUsers(q), placeholderData: keepPreviousData });
}

export function useAdminFeedback(q: TAdminFeedbackQuery) {
  return useQuery({
    queryKey: adminKeys.feedbackList(q),
    queryFn: () => fetchAdminFeedback(q),
    placeholderData: keepPreviousData,
  });
}

export function useUpdateFeedback() {
  const qc = useQueryClient();
  return useMutation<TAdminFeedback, ApiError, { id: string } & TUpdateFeedbackRequest>({
    mutationFn: ({ id, ...body }) => updateFeedbackApi(id, body),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: adminKeys.feedback() });
      void qc.invalidateQueries({ queryKey: adminKeys.overview() });
    },
  });
}

export function useAdminCandidates() {
  return useQuery({ queryKey: adminKeys.candidates(), queryFn: fetchAdminCandidates });
}

/** Approve or reject library candidates (Epic 9); the list and the learners' library refresh. */
export function useDecideCandidates() {
  const qc = useQueryClient();
  const onSuccess = () => {
    void qc.invalidateQueries({ queryKey: adminKeys.candidates() });
    void qc.invalidateQueries({ queryKey: ["library"] });
  };
  const approve = useMutation<{ count: number }, ApiError, TApproveCandidatesRequest>({ mutationFn: approveCandidatesApi, onSuccess });
  const reject = useMutation<{ count: number }, ApiError, string[]>({ mutationFn: rejectCandidatesApi, onSuccess });
  return { approve, reject, pending: approve.isPending || reject.isPending };
}
