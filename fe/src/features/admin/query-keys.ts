import type { TAdminFeedbackQuery, TAdminUsersQuery } from "./types";

export const adminKeys = {
  all: ["admin"] as const,
  overview: () => [...adminKeys.all, "overview"] as const,
  users: (q: TAdminUsersQuery) => [...adminKeys.all, "users", q] as const,
  feedback: () => [...adminKeys.all, "feedback"] as const,
  feedbackList: (q: TAdminFeedbackQuery) => [...adminKeys.feedback(), q] as const,
};
