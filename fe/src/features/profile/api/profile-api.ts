import { apiClient } from "@/lib/api-client";
import type {
  TMe,
  TStudySettings,
  TUpdateProfilePayload,
  TUpdateStudySettingsPayload,
} from "../types";

export const fetchMe = () => apiClient.get<TMe>("/me");

export const updateProfileApi = (payload: TUpdateProfilePayload) =>
  apiClient.patch<TMe>("/me", payload);

export const updateStudySettingsApi = (payload: TUpdateStudySettingsPayload) =>
  apiClient.patch<TStudySettings>("/me/study-settings", payload);
