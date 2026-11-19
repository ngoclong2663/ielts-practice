import type { TCardContent } from "@/types/card-content";

// Vocabulary library (Epic 8 §3): reviewed cards added to a deck without the AI.

export type TLocalized = { vi: string; en: string };

export type TLibraryCollection = {
  slug: string;
  kind: string;
  title: TLocalized;
  description: TLocalized;
  words: number;
  topics: { id: string; words: number }[];
};

export type TLibraryLevel = "A1" | "A2" | "B1" | "B2" | "C1" | "C2";

export type TLibraryWord = {
  id: string;
  topic: string;
  lexeme_id: string;
  cefr: TLibraryLevel | "";
  source: "ielts4000" | "core";
  card: TCardContent;
  in_deck: boolean;
  known: boolean;
};

export type TLibraryWords = {
  collection: TLibraryCollection;
  words: TLibraryWord[];
  truncated: boolean;
};

export type TLibraryWordsParams = { topic?: string; q?: string };

export type TAddLibraryWordsPayload = { lexeme_ids: string[]; deck_id?: string; new_deck_name?: string };
export type TAddLibraryWordsResult = { deck_id: string; deck_name: string; created: number; skipped: string[] };
