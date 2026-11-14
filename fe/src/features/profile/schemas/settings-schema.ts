import { z } from "zod";

// Error messages are translation keys under `validation.*`.
export const profileFormSchema = z.object({
  display_name: z.string().trim().min(1, "required").max(120, "tooLong"),
  ui_locale: z.enum(["vi", "en"]),
  timezone: z.string().min(1, "required"),
  day_start_hour: z.number().int().min(0, "range").max(23, "range"),
});

export type TProfileFormValues = z.infer<typeof profileFormSchema>;

export const studySettingsFormSchema = z.object({
  new_cards_per_day: z.number({ error: "range" }).int("range").min(0, "range").max(500, "range"),
  max_reviews_per_day: z.number({ error: "range" }).int("range").min(0, "range").max(9999, "range"),
  desired_retention: z.number().min(0.8).max(0.97),
  audio_autoplay: z.boolean(),
  accent: z.enum(["uk", "us"]),
  show_image_on_front: z.boolean(),
});

const STEPS = /^\s*(\d+(\.\d+)?[mhd](\s+\d+(\.\d+)?[mhd]){0,5})?\s*$/;

export const advancedSettingsFormSchema = z.object({
  learning_steps: z.string().regex(STEPS, "steps"),
  relearning_steps: z.string().regex(STEPS, "steps"),
  max_interval_days: z.number({ error: "range" }).int("range").min(1, "range").max(36500, "range"),
});

export type TAdvancedSettingsFormValues = z.infer<typeof advancedSettingsFormSchema>;

export type TStudySettingsFormValues = z.infer<typeof studySettingsFormSchema>;
