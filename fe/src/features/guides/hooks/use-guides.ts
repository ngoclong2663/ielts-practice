import { useQuery } from "@tanstack/react-query";
import { useLocale } from "next-intl";
import { fetchGuides } from "../api/guides-api";
import { guideKeys } from "../query-keys";
import type { TGuide, TGuideTopic, TLocalized } from "../types";

/** The guides of one topic. They change only with a deploy, so they are fetched once per session. */
export function useGuides(topic: TGuideTopic) {
  return useQuery({ queryKey: guideKeys.list(topic), queryFn: () => fetchGuides(topic), staleTime: Infinity, retry: 1 });
}

/** The first guide of a topic, or undefined while loading or when there is none. */
export function useGuide(topic: TGuideTopic): TGuide | undefined {
  return useGuides(topic).data?.[0];
}

/** Picks the text for the UI language (Vietnamese when a translation is missing). */
export function useLocalized() {
  const locale = useLocale();
  return (text: TLocalized) => (locale === "en" ? text.en || text.vi : text.vi);
}

/** "0:52" */
export function formatDuration(seconds: number): string {
  const s = Math.round(seconds);
  return `${Math.floor(s / 60)}:${String(s % 60).padStart(2, "0")}`;
}
