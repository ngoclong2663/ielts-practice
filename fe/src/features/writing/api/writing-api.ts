import { apiClient } from "@/lib/api-client";
import type { TAttempt, TAttemptSummary, TLevel, TPosition, TTaskDetail, TTaskSummary, TWritingStatus } from "../types";

export const fetchWritingStatus = () => apiClient.get<TWritingStatus>("/writing/status");
export const fetchTasks = (taskNo: number) => apiClient.get<TTaskSummary[]>(`/writing/tasks?task_no=${taskNo}`);
export const fetchTask = (id: string) => apiClient.get<TTaskDetail>(`/writing/tasks/${id}`);
export const fetchAttempts = () => apiClient.get<TAttemptSummary[]>("/writing/attempts");
export const fetchAttempt = (id: string) => apiClient.get<TAttempt>(`/writing/attempts/${id}`);

export const startAttemptApi = (taskId: string, p: { level: TLevel; position?: TPosition }) =>
  apiClient.post<TAttempt>(`/writing/tasks/${taskId}/attempts`, p);
export const answerPlanApi = (id: string, answers: number[]) =>
  apiClient.post<TAttempt>(`/writing/attempts/${id}/plan`, { answers });
export const checkParagraphApi = (id: string, key: string, answers: string[]) =>
  apiClient.post<TAttempt>(`/writing/attempts/${id}/paragraphs/${key}/check`, { answers });
export const submitAttemptApi = (id: string, p: { text: string; seconds_spent: number }) =>
  apiClient.post<TAttempt>(`/writing/attempts/${id}/submit`, p);
