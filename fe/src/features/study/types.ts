import type { TCardContent, TCardState } from "@/types/card-content";

export type TCounts = { new: number; learning: number; review: number };

export type TStudyCard = {
  id: string;
  deck_id: string;
  deck_name: string;
  state: TCardState;
  content: TCardContent;
  is_leech: boolean;
  reps: number;
  lapses: number;
  intervals: { again: number; hard: number; good: number; easy: number };
};

export type TStudyQueue = {
  cards: TStudyCard[];
  counts: TCounts;
  next_due_at: string | null;
  day_ends_at: string;
};

export type TRating = 1 | 2 | 3 | 4;

export type TAnswerPayload = {
  card_id: string;
  rating: TRating;
  duration_ms: number;
  client_review_id: string;
  expected_reps: number;
  deck_id?: string;
  size?: number;
};

export type TAnswerResponse = TStudyQueue & {
  answered: { card_id: string; rating: number; state: TCardState; due: string; became_leech: boolean } | null;
};

export type TUndoResponse = TStudyQueue & { undone_card_id: string };

export type TCountsResponse = {
  decks: (TCounts & { deck_id: string })[];
  total: TCounts;
  day_ends_at: string;
};
