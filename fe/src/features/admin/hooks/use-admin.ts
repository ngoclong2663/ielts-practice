import { keepPreviousData, useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import type { ApiError } from "@/lib/api-client";
import { fetchAdminFeedback, fetchAdminOverview, fetchAdminUsers, updateFeedbackApi } from "../api/admin-api";
import { adminKeys } from "../query-keys";
import type { TAdminFeedback, TAdminFeedbackQuery, TAdminUsersQuery, TUpdateFeedbackRequest } from "../types";

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
