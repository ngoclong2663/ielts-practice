import type { TCardContent, TCardImage } from "@/types/card-content";

export type TEntry = { text: string; sense_hint: string };

export type TRejectReason = "TOO_LONG" | "TOO_MANY_WORDS" | "INVALID_CHARS" | "DUPLICATE";
export type TRejectedEntry = { text: string; reason: TRejectReason };

export type TItemStatus = "queued" | "content" | "image" | "ready" | "failed" | "invalid";
export type TJobStatus = "queued" | "running" | "done" | "partial" | "failed";

export type TJobItem = {
  id: string;
  position: number;
  input_text: string;
  sense_hint: string;
  status: TItemStatus;
  error_code: string;
  suggestion: string;
  cached: boolean;
  lexeme_id: string | null;
  content: TCardContent | null;
  image_candidates: TCardImage[];
  removed: boolean;
};

export type TJob = {
  id: string;
  status: TJobStatus;
  total: number;
  done: number;
  target_deck_id: string | null;
  committed_deck_id: string | null;
  created_at: string;
  finished_at: string | null;
  items: TJobItem[];
};

export type TJobSummary = Omit<TJob, "items" | "target_deck_id"> & { words: string[] };

export type TVocabStatus = {
  configured: boolean;
  max_entries: number;
  quota: { limit: number; used: number; remaining: number };
  suggestions: { limit: number; used: number; remaining: number };
  image_providers: string[];
};

export type TImageSearchResult = { images: TCardImage[]; provider: string; providers: string[] };

export type TUpdateItemPayload = {
  content?: TCardContent;
  image?: { provider: string; provider_id: string };
  clear_image?: boolean;
  removed?: boolean;
};

export type TRetryPayload = { input_text?: string; sense_hint?: string; regenerate?: boolean };

export type TCommitPayload = { deck_id?: string; new_deck_name?: string; item_ids: string[] };
export type TCommitResult = { deck_id: string; deck_name: string; created: number; skipped: string[] };

export type TReportReason = "wrong_meaning" | "bad_example" | "wrong_image" | "other";

export const FINAL_ITEM_STATUSES: readonly TItemStatus[] = ["ready", "failed", "invalid"];
export const FINAL_JOB_STATUSES: readonly TJobStatus[] = ["done", "partial", "failed"];

// ---- words by topic (Epic 4)

export type TSuggestedWord = { lemma: string; pos: string; meaning_vi: string; cefr: string };

export type TSuggestion = {
  id: string;
  topic_id: string;
  topic_text: string;
  deck_id: string | null;
  requested: number;
  words: TSuggestedWord[];
  skipped_existing: number;
  added: number;
  created_at: string;
  remaining_requests: number;
};

export type TCreateSuggestionPayload = {
  topic_id?: string;
  topic_text?: string;
  count: number;
  deck_id?: string;
};
