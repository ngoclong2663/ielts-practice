export type TGuideTopic = "vocab" | "writing";
export type TLocalized = { vi: string; en: string };

/** A short guide video (Epic 7 §5). URLs are API paths that redirect to a short-lived S3 link. */
export type TGuide = {
  slug: string;
  topic: TGuideTopic;
  /** Language of the narration. */
  lang: string;
  title: TLocalized;
  description: TLocalized;
  duration_seconds: number;
  video_url: string;
  poster_url: string;
};
