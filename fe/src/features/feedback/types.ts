export const FEEDBACK_CATEGORIES = ["bug", "idea", "content", "other"] as const;
export type TFeedbackCategory = (typeof FEEDBACK_CATEGORIES)[number];

export const FEEDBACK_MIN_LENGTH = 10;
export const FEEDBACK_MAX_LENGTH = 2000;

export type TSendFeedbackRequest = {
  category: TFeedbackCategory;
  message: string;
  page?: string;
  ui_locale?: "vi" | "en";
};

export type TSendFeedbackResponse = { id: string; created_at: string };
