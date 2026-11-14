import { useMutation, useQueryClient } from "@tanstack/react-query";
import type { ApiError } from "@/lib/api-client";
import { updateProfileApi, updateStudySettingsApi } from "../api/profile-api";
import { profileKeys } from "../query-keys";
import type { TMe, TStudySettings, TUpdateProfilePayload, TUpdateStudySettingsPayload } from "../types";

export function useUpdateProfile() {
  const queryClient = useQueryClient();
  return useMutation<TMe, ApiError, TUpdateProfilePayload>({
    mutationFn: updateProfileApi,
    onSuccess: (me) => {
      queryClient.setQueryData(profileKeys.me(), me);
    },
  });
}

export function useUpdateStudySettings() {
  const queryClient = useQueryClient();
  return useMutation<TStudySettings, ApiError, TUpdateStudySettingsPayload>({
    mutationFn: updateStudySettingsApi,
    onSuccess: (settings) => {
      queryClient.setQueryData<TMe>(profileKeys.me(), (prev) =>
        prev ? { ...prev, study_settings: settings } : prev,
      );
    },
  });
}
