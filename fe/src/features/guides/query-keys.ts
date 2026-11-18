import type { TGuideTopic } from "./types";

export const guideKeys = {
  all: ["guides"] as const,
  list: (topic: TGuideTopic) => [...guideKeys.all, "list", topic] as const,
};
