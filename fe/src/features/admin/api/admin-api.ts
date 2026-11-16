import type {
  TAdminFeedback,
  TAdminFeedbackPage,
  TAdminFeedbackQuery,
  TAdminOverview,
  TAdminUser,
  TAdminUsersQuery,
  TPage,
  TUpdateFeedbackRequest,
} from "../types";
import { adminClient } from "./admin-client";

const PAGE_SIZE = 25;

function qs(params: Record<string, string | number>): string {
  const search = new URLSearchParams();
  for (const [key, value] of Object.entries(params)) {
    if (value !== "") search.set(key, String(value));
  }
  return search.toString();
}

export const fetchAdminOverview = () => adminClient.get<TAdminOverview>("/overview");

export const fetchAdminUsers = (q: TAdminUsersQuery) =>
  adminClient.get<TPage<TAdminUser>>(`/users?${qs({ q: q.q.trim(), sort: q.sort, page: q.page, page_size: PAGE_SIZE })}`);

export const fetchAdminFeedback = (q: TAdminFeedbackQuery) =>
  adminClient.get<TAdminFeedbackPage>(
    `/feedback?${qs({ status: q.status, category: q.category, page: q.page, page_size: PAGE_SIZE })}`,
  );

export const updateFeedbackApi = (id: string, body: TUpdateFeedbackRequest) =>
  adminClient.patch<TAdminFeedback>(`/feedback/${id}`, body);
