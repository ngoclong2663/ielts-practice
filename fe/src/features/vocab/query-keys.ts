export const vocabKeys = {
  all: ["vocab"] as const,
  status: () => [...vocabKeys.all, "status"] as const,
  jobs: () => [...vocabKeys.all, "jobs"] as const,
  job: (id: string) => [...vocabKeys.all, "job", id] as const,
  topics: () => [...vocabKeys.all, "topics"] as const,
  suggestion: (id: string) => [...vocabKeys.all, "suggestion", id] as const,
};
