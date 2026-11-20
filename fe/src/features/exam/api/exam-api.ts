import { apiClient } from "@/lib/api-client";
import type { TAttempt, TAttemptList, TBook, TSkill, TStats, TSubmitAttempt, TTest } from "../types";

export const fetchBooks = (skill: TSkill) => apiClient.get<TBook[]>(`/exam/books?skill=${skill}`);
export const fetchTest = (slug: string, skill: TSkill) =>
  apiClient.get<TTest>(`/exam/tests/${encodeURIComponent(slug)}?skill=${skill}`);
export const fetchAttempt = (id: string) => apiClient.get<TAttempt>(`/exam/attempts/${id}`);
export const fetchAttempts = (skill: TSkill, before?: string) =>
  apiClient.get<TAttemptList>(`/exam/attempts?skill=${skill}${before ? `&before=${before}` : ""}`);
export const fetchStats = (skill: TSkill) => apiClient.get<TStats>(`/exam/stats?skill=${skill}`);
export const submitAttemptApi = (body: TSubmitAttempt) => apiClient.post<{ id: string }>("/exam/attempts", body);
