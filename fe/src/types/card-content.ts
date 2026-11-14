// Mirrors the backend card content JSON (schema v1, Epic 3 §4).
export type TExample = { en: string; vi?: string };

export type TCardContent = {
  headword: string;
  lemma?: string;
  ipa?: { uk?: string; us?: string };
  part_of_speech?: string;
  cefr?: string;
  meaning_vi: string;
  definition_en?: string;
  word_forms?: { form: string; pos?: string; ipa?: string; meaning_vi?: string; example?: TExample }[];
  collocations?: { phrase: string; meaning_vi?: string }[];
  extra_examples?: TExample[];
  synonyms?: string[];
  other_senses?: { pos?: string; meaning_vi: string }[];
  image?: {
    provider: string;
    url: string;
    thumb_url?: string;
    width?: number;
    height?: number;
    alt?: string;
    author_name?: string;
    author_url?: string;
    source_url?: string;
  };
  audio?: { headword?: Partial<Record<"uk" | "us", string>>; examples?: Record<string, string> };
};

export type TCardState = "new" | "learning" | "review" | "relearning";
