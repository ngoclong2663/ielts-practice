export type TUiLocale = "vi" | "en";
export type TAccent = "uk" | "us";

export type TUserProfile = {
  id: string;
  email: string;
  display_name: string;
  avatar_url: string;
  ui_locale: TUiLocale;
  timezone: string;
  day_start_hour: number;
  created_at: string;
};

export type TStudySettings = {
  new_cards_per_day: number;
  max_reviews_per_day: number;
  desired_retention: number;
  learning_steps: string;
  relearning_steps: string;
  audio_autoplay: boolean;
  accent: TAccent;
  show_image_on_front: boolean;
  max_interval_days: number;
};

export type TMe = {
  user: TUserProfile;
  study_settings: TStudySettings;
};

export type TUpdateProfilePayload = Partial<
  Pick<TUserProfile, "display_name" | "ui_locale" | "timezone" | "day_start_hour">
>;

export type TUpdateStudySettingsPayload = Partial<TStudySettings>;
