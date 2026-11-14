import { z } from "zod";
import type { TCardContent } from "@/types/card-content";

// Messages are translation keys under `validation.*`.
const text = (max: number) => z.string().trim().max(max, "tooLong");

export const contentFormSchema = z.object({
  headword: z.string().trim().min(1, "required").max(100, "tooLong"),
  ipa_uk: text(100),
  ipa_us: text(100),
  part_of_speech: text(30),
  meaning_vi: z.string().trim().min(1, "required").max(300, "tooLong"),
  definition_en: text(300),
  word_forms: z.array(
    z.object({
      form: z.string().trim().min(1, "required").max(100, "tooLong"),
      pos: text(30),
      ipa: text(100),
      meaning_vi: text(300),
      example_en: text(300),
      example_vi: text(300),
    }),
  ),
  collocations: z.array(
    z.object({ phrase: z.string().trim().min(1, "required").max(100, "tooLong"), meaning_vi: text(300) }),
  ),
  extra_examples: z.array(z.object({ en: z.string().trim().min(1, "required").max(300, "tooLong"), vi: text(300) })),
  synonyms: text(300),
});

export type TContentFormValues = z.infer<typeof contentFormSchema>;

export function contentToForm(c: TCardContent): TContentFormValues {
  return {
    headword: c.headword ?? "",
    ipa_uk: c.ipa?.uk ?? "",
    ipa_us: c.ipa?.us ?? "",
    part_of_speech: c.part_of_speech ?? "",
    meaning_vi: c.meaning_vi ?? "",
    definition_en: c.definition_en ?? "",
    word_forms: (c.word_forms ?? []).map((w) => ({
      form: w.form,
      pos: w.pos ?? "",
      ipa: w.ipa ?? "",
      meaning_vi: w.meaning_vi ?? "",
      example_en: w.example?.en ?? "",
      example_vi: w.example?.vi ?? "",
    })),
    collocations: (c.collocations ?? []).map((x) => ({ phrase: x.phrase, meaning_vi: x.meaning_vi ?? "" })),
    extra_examples: (c.extra_examples ?? []).map((x) => ({ en: x.en, vi: x.vi ?? "" })),
    synonyms: (c.synonyms ?? []).join(", "),
  };
}

/** Applies the form onto the original content, keeping fields the form doesn't show (image, CEFR, senses). */
export function formToContent(v: TContentFormValues, original: TCardContent): TCardContent {
  const ipa = v.ipa_uk || v.ipa_us ? { uk: v.ipa_uk, us: v.ipa_us } : undefined;
  return {
    ...original,
    headword: v.headword,
    ipa,
    part_of_speech: v.part_of_speech,
    meaning_vi: v.meaning_vi,
    definition_en: v.definition_en,
    word_forms: v.word_forms.map((w) => ({
      form: w.form,
      pos: w.pos,
      ipa: w.ipa,
      meaning_vi: w.meaning_vi,
      example: w.example_en ? { en: w.example_en, vi: w.example_vi } : undefined,
    })),
    collocations: v.collocations,
    extra_examples: v.extra_examples,
    synonyms: v.synonyms
      .split(",")
      .map((s) => s.trim())
      .filter(Boolean),
  };
}
