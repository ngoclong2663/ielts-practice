export const studyKeys = {
  all: ["study"] as const,
  counts: () => [...studyKeys.all, "counts"] as const,
  queue: (deckId: string) => [...studyKeys.all, "queue", deckId || "all"] as const,
  stats: () => [...studyKeys.all, "stats"] as const,
};
