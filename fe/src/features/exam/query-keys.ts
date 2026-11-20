import type { TSkill } from "./types";

export const examKeys = {
  all: ["exam"] as const,
  books: (skill: TSkill) => [...examKeys.all, "books", skill] as const,
  test: (slug: string, skill: TSkill) => [...examKeys.all, "test", slug, skill] as const,
  attempts: (skill: TSkill) => [...examKeys.all, "attempts", skill] as const,
  attempt: (id: string) => [...examKeys.all, "attempt", id] as const,
  stats: (skill: TSkill) => [...examKeys.all, "stats", skill] as const,
};
