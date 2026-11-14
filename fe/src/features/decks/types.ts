import type { TCardContent, TCardState } from "@/types/card-content";

export type TDeck = {
  id: string;
  name: string;
  description: string;
  source: "manual" | "paste" | "collection";
  collection_id: string | null;
  new_per_day: number | null;
  reviews_per_day: number | null;
  card_count: number;
  created_at: string;
};

export type TCardSummary = {
  id: string;
  deck_id: string;
  headword: string;
  meaning_vi: string;
  state: TCardState;
  due: string;
  scheduled_days: number;
  reps: number;
  lapses: number;
  suspended: boolean;
  is_leech: boolean;
  created_at: string;
};

export type TCard = TCardSummary & { content: TCardContent };

export type TCardList = { items: TCardSummary[]; page: number; page_size: number; total: number };

export type TCardStateFilter = "" | "new" | "learning" | "review" | "suspended" | "leech";

export type TCardListParams = { q: string; state: TCardStateFilter; page: number };

export type TCreateDeckPayload = { name: string; description?: string };

export type TUpdateDeckPayload = {
  name?: string;
  description?: string;
  new_per_day?: number;
  reviews_per_day?: number;
  use_default_limits?: boolean;
};

export type TManualCardFields = {
  headword?: string;
  meaning_vi?: string;
  ipa?: string;
  part_of_speech?: string;
  example_en?: string;
  example_vi?: string;
  image_url?: string;
};

export type TBulkAction = "suspend" | "unsuspend" | "reset" | "delete" | "move";
