import { z } from "zod";

// Messages are translation keys under `validation.*`.
export const deckFormSchema = z
  .object({
    name: z.string().trim().min(1, "required").max(80, "tooLong"),
    description: z.string().trim().max(500, "tooLong"),
    use_default_limits: z.boolean(),
    new_per_day: z.number({ error: "range" }).int("range").min(0, "range").max(500, "range"),
    reviews_per_day: z.number({ error: "range" }).int("range").min(0, "range").max(9999, "range"),
  });

export type TDeckFormValues = z.infer<typeof deckFormSchema>;

const optionalText = (max: number) => z.string().trim().max(max, "tooLong");

export const cardFormSchema = z.object({
  headword: z.string().trim().min(1, "required").max(100, "tooLong"),
  meaning_vi: z.string().trim().min(1, "required").max(300, "tooLong"),
  ipa: optionalText(100),
  part_of_speech: optionalText(30),
  example_en: optionalText(300),
  example_vi: optionalText(300),
  image_url: z.union([z.literal(""), z.url("url").max(1000, "tooLong")]),
});

export type TCardFormValues = z.infer<typeof cardFormSchema>;
