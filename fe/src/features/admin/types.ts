import type { TFeedbackCategory } from "@/features/feedback/types";

// Mirrors GET /api/v1/admin/* (Epic 6 §8.1).

export type TAdminDaily = { date: string; signups: number; active_users: number; reviews: number; ai_requests: number };

export type TAdminReport = {
  id: string;
  headword: string;
  reason: string;
  note: string;
  user_email: string;
  created_at: string;
};

export type TAdminOverview = {
  generated_at: string;
  timezone: string;
  users: { total: number; new_7d: number; new_30d: number; active_1d: number; active_7d: number; active_30d: number };
  vocab: {
    decks: number;
    cards: number;
    cards_by_state: { new: number; learning: number; review: number; relearning: number };
    reviews_total: number;
    reviews_7d: number;
    lexemes_cached: number;
  };
  ai: {
    words_generated_total: number;
    words_generated_7d: number;
    words_cached_total: number;
    words_failed_7d: number;
    topic_lists_total: number;
    topic_lists_7d: number;
    writing_lessons_total: number;
    writing_checks_total: number;
    writing_checks_7d: number;
    writing_reviews_total: number;
    writing_reviews_7d: number;
    input_tokens_30d: number;
    output_tokens_30d: number;
  };
  writing: { attempts_total: number; attempts_submitted: number; avg_band_30d: number | null };
  feedback: { new: number; total: number };
  content_reports: { total: number; last_7d: number; needs_review: number };
  daily: TAdminDaily[];
  recent_reports: TAdminReport[];
};

export type TPage<TItem> = { items: TItem[]; total: number; page: number; page_size: number };

export type TAdminUser = {
  id: string;
  email: string;
  display_name: string;
  avatar_url: string;
  ui_locale: "vi" | "en";
  created_at: string;
  last_login_at: string | null;
  last_active_at: string | null;
  decks: number;
  cards: number;
  reviews_total: number;
  reviews_7d: number;
  ai_words: number;
  topic_lists: number;
  writing_attempts: number;
  writing_submitted: number;
  writing_ai_calls: number;
  feedback: number;
};

export const ADMIN_USER_SORTS = ["last_active", "created", "reviews", "ai"] as const;
export type TAdminUserSort = (typeof ADMIN_USER_SORTS)[number];
export type TAdminUsersQuery = { q: string; sort: TAdminUserSort; page: number };

export const FEEDBACK_STATUSES = ["new", "seen", "resolved"] as const;
export type TFeedbackStatus = (typeof FEEDBACK_STATUSES)[number];
export type TFeedbackStatusFilter = TFeedbackStatus | "all";

export type TAdminFeedback = {
  id: string;
  category: TFeedbackCategory;
  message: string;
  page: string;
  ui_locale: string;
  user_agent: string;
  status: TFeedbackStatus;
  admin_note: string;
  notified: boolean;
  created_at: string;
  updated_at: string;
  resolved_at: string | null;
  user: { id: string; email: string; display_name: string; avatar_url: string };
};

export type TAdminFeedbackPage = TPage<TAdminFeedback> & { counts: Record<TFeedbackStatus, number> };
export type TAdminFeedbackQuery = { status: TFeedbackStatusFilter; category: TFeedbackCategory | ""; page: number };
export type TUpdateFeedbackRequest = { status?: TFeedbackStatus; admin_note?: string };
